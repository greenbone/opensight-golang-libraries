// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package osquery

import (
	"fmt"
	"reflect"

	"github.com/aquasecurity/esquery"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
)

// BoolQueryBuilder is a builder for an OpenSearch bool query.
// Use NewBoolQueryBuilder or NewBoolQueryBuilderWith for proper initialization.
type BoolQueryBuilder struct {
	querySettings    *QuerySettings
	compareOperators []CompareOperator
	query            *esquery.BoolQuery
	Must             []esquery.Mappable
	MustNot          []esquery.Mappable
}

type (
	queryAppender func(fieldName string, fieldKeys []string, fieldValue any) error
	// CompareOperatorHandler is a function that generates an appropriate query condition for the given field.
	CompareOperatorHandler func(fieldName string, fieldValue any) (esquery.Mappable, error)
)

// QuerySettings is used to configure the query builder.
type QuerySettings struct {
	FilterFieldMapping map[string]string
}

// CompareOperator defines a mapping between a filter.CompareOperator and a function to generate an appropriate
// query condition in form of a CompareOperatorHandler.
type CompareOperator struct {
	Operator filter.CompareOperator
	Handler  CompareOperatorHandler
	// MustCondition defines whether the condition should be added to the must (true) or must_not clause (false).
	MustCondition bool
}

// NewBoolQueryBuilder creates a new BoolQueryBuilder and returns it. It uses the default set of CompareOperator.
//
// querySettings is used to configure the query builder.
func NewBoolQueryBuilder(querySettings *QuerySettings) *BoolQueryBuilder {
	return NewBoolQueryBuilderWith(esquery.Bool(), querySettings)
}

// NewBoolQueryBuilderWith creates a new BoolQueryBuilder and returns it. It uses the default set of CompareOperator.
//
// query is the initial bool query to use.
// querySettings is used to configure the query builder.
func NewBoolQueryBuilderWith(query *esquery.BoolQuery, querySettings *QuerySettings) *BoolQueryBuilder {
	return &BoolQueryBuilder{
		querySettings:    querySettings,
		compareOperators: defaultCompareOperators(),
		query:            query,
	}
}

// AddTermsFilter adds a terms filter to this query.
//
// values is the list of values to filter for.
func (q *BoolQueryBuilder) AddTermsFilter(fieldName string, values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("need at least one value for terms filter")
	}

	entityName, ok := q.querySettings.FilterFieldMapping[fieldName]
	if !ok {
		return fmt.Errorf("Mapping for filter field '%s' is currently not implemented.", fieldName)
	}

	q.query = q.query.Filter(esquery.Terms(entityName, values...))
	return nil
}

// AddTermFilter adds a term filter to this query.
//
// value is the value to filter for.
func (q *BoolQueryBuilder) AddTermFilter(fieldName string, value any) error {
	entityName, ok := q.querySettings.FilterFieldMapping[fieldName]
	if !ok {
		return fmt.Errorf("Mapping for filter field '%s' is currently not implemented.", fieldName)
	}

	q.query = q.query.Filter(esquery.Term(entityName, value))
	return nil
}

func (q *BoolQueryBuilder) addToMust(call CompareOperatorHandler) queryAppender {
	return func(fieldName string, fieldKeys []string, fieldValue any) error {
		value, err := call(fieldName, fieldValue)
		if err != nil {
			return err
		}
		if value != nil {
			q.Must = append(q.Must, value)
		}
		return nil
	}
}

func (q *BoolQueryBuilder) addToMustNot(call CompareOperatorHandler) queryAppender {
	return func(fieldName string, fieldKeys []string, fieldValue any) error {
		value, err := call(fieldName, fieldValue)
		if err != nil {
			return err
		}
		if value != nil {
			q.MustNot = append(q.MustNot, value)
		}
		return nil
	}
}

func (q *BoolQueryBuilder) createOperatorMapping() map[filter.CompareOperator]queryAppender {
	operatorMapping := make(map[filter.CompareOperator]queryAppender,
		len(q.compareOperators))
	for _, setting := range q.compareOperators {
		if setting.MustCondition {
			operatorMapping[setting.Operator] = q.addToMust(setting.Handler)
		} else {
			operatorMapping[setting.Operator] = q.addToMustNot(setting.Handler)
		}
	}
	return operatorMapping
}

// AddFilterRequest adds a filter request to this query.
// The filter request is translated into a bool query.
func (q *BoolQueryBuilder) AddFilterRequest(request *filter.Request) error {
	if request == nil || len(request.Fields) == 0 {
		return nil
	}
	if request.Operator == "" && len(request.Fields) == 1 { // for single filter `Operator` is not relevant
		request.Operator = filter.LogicOperatorAnd
	}

	effectiveRequest, err := effectiveFilterFields(*request, q.querySettings.FilterFieldMapping)
	if err != nil {
		return err
	}

	operatorMapping := q.createOperatorMapping()

	for _, field := range effectiveRequest.Fields {
		if handler, ok := operatorMapping[field.Operator]; ok {
			value := field.Value

			if field.Operator == filter.CompareOperatorExists {
				value = "" // exists operator does not need a value, but for more consistent handling just pass a dummy value
			}
			if value == nil {
				return fmt.Errorf("field '%s' has no value set", field.Name)
			}

			if t := reflect.TypeOf(value); t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
				slice := reflect.ValueOf(value)

				if slice.Len() == 0 { // disallow empty list values, as the there is no clear way to interpret this kind of filter
					return fmt.Errorf("field '%s' has empty list of values", field.Name)
				}
				// convert to []any, so that handlers don't need to deal with different slice types
				var values []any
				if v, ok := value.([]any); ok {
					values = v
				} else {
					values = make([]any, slice.Len())
					for i := 0; i < slice.Len(); i++ {
						values[i] = slice.Index(i).Interface()
					}
				}
				value = values
			}

			err := handler(field.Name, field.Keys, value)
			if err != nil {
				return fmt.Errorf("failed to transform filter with operator %q to database query: %w", field.Operator, err)
			}
		} else {
			return fmt.Errorf("field '%s' with unknown operator '%s'", field.Name, field.Operator)
		}
	}
	switch effectiveRequest.Operator {
	case filter.LogicOperatorAnd:
		q.query = q.query.
			Must(q.Must...).
			MustNot(q.MustNot...)
		return nil
	case filter.LogicOperatorOr:
		if len(q.Must) > 0 || len(q.MustNot) > 0 {
			shouldQueries := make([]esquery.Mappable, 0, len(q.Must)+len(q.MustNot))
			shouldQueries = append(shouldQueries, q.Must...)
			// For each MustNot condition, create a new bool query with a must_not clause
			// and add it to shouldQueries
			for _, mustNotQuery := range q.MustNot {
				negatedQuery := esquery.Bool().
					MustNot(mustNotQuery)
				shouldQueries = append(shouldQueries, negatedQuery)
			}
			q.query = q.query.Should(shouldQueries...).MinimumShouldMatch(1)
		}
		return nil
	default:
		return fmt.Errorf("unknown operator '%s'", effectiveRequest.Operator)
	}
}

func effectiveFilterFields(filterRequest filter.Request, fieldMapping map[string]string) (filter.Request, error) {
	var filterFields []filter.RequestField
	for _, field := range filterRequest.Fields {
		mappedField, err := createMappedField(field, fieldMapping)
		if err != nil {
			return filter.Request{}, err
		}
		filterFields = append(filterFields, mappedField)
	}
	return filter.Request{
		Operator: filterRequest.Operator,
		Fields:   filterFields,
	}, nil
}

func createMappedField(dtoField filter.RequestField, fieldMapping map[string]string) (filter.RequestField, error) {
	entityName, ok := fieldMapping[dtoField.Name]
	if !ok {
		return filter.RequestField{}, filter.NewInvalidFilterFieldError(
			"Mapping for filter field '%s' is currently not implemented.", dtoField.Name)
	}

	return filter.RequestField{
		Operator: dtoField.Operator,
		Keys:     dtoField.Keys,
		Name:     entityName,
		Value:    dtoField.Value,
	}, nil
}

// Build returns the built query.
func (q *BoolQueryBuilder) Build() *esquery.BoolQuery {
	return q.query
}

func defaultCompareOperators() []CompareOperator {
	return []CompareOperator{
		{Operator: filter.CompareOperatorIsEqualTo, Handler: HandleCompareOperatorIsEqualTo, MustCondition: true},
		{Operator: filter.CompareOperatorIsNotEqualTo, Handler: HandleCompareOperatorIsEqualTo, MustCondition: false},
		{Operator: filter.CompareOperatorIsNumberEqualTo, Handler: HandleCompareOperatorIsEqualTo, MustCondition: true},
		{
			Operator: filter.CompareOperatorIsNumberNotEqualTo,
			Handler:  HandleCompareOperatorIsEqualTo, MustCondition: false,
		},
		{Operator: filter.CompareOperatorIsIpEqualTo, Handler: HandleCompareOperatorIsEqualTo, MustCondition: true},
		{Operator: filter.CompareOperatorIsIpNotEqualTo, Handler: HandleCompareOperatorIsEqualTo, MustCondition: false},
		{Operator: filter.CompareOperatorIsStringEqualTo, Handler: HandleCompareOperatorIsEqualTo, MustCondition: true},
		{
			Operator: filter.CompareOperatorIsStringNotEqualTo,
			Handler:  HandleCompareOperatorIsEqualTo, MustCondition: false,
		},
		{Operator: filter.CompareOperatorContains, Handler: HandleCompareOperatorContains, MustCondition: true},
		{Operator: filter.CompareOperatorDoesNotContain, Handler: HandleCompareOperatorContains, MustCondition: false},
		{Operator: filter.CompareOperatorBeginsWith, Handler: HandleCompareOperatorBeginsWith, MustCondition: true},
		{
			Operator: filter.CompareOperatorDoesNotBeginWith,
			Handler:  HandleCompareOperatorBeginsWith, MustCondition: false,
		},
		{
			Operator: filter.CompareOperatorIsLessThanOrEqualTo,
			Handler:  HandleCompareOperatorIsLessThanOrEqualTo, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
			Handler:  HandleCompareOperatorIsGreaterThanOrEqualTo, MustCondition: true,
		},
		{Operator: filter.CompareOperatorTextContains, Handler: HandleCompareOperatorTextContains, MustCondition: true},
		{Operator: filter.CompareOperatorIsGreaterThan, Handler: HandleCompareOperatorIsGreaterThan, MustCondition: true},
		{Operator: filter.CompareOperatorIsLessThan, Handler: HandleCompareOperatorIsLessThan, MustCondition: true},
		{Operator: filter.CompareOperatorAfterDate, Handler: HandleCompareOperatorIsGreaterThan, MustCondition: true},
		{Operator: filter.CompareOperatorBeforeDate, Handler: HandleCompareOperatorIsLessThan, MustCondition: true},
		{
			Operator: filter.CompareOperatorBetweenDates,
			Handler:  HandleCompareOperatorBetweenDates, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorExists,
			Handler:  HandleCompareOperatorExists, MustCondition: true,
		},
	}
}
