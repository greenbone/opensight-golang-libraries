// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"fmt"

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
	queryAppender func(fieldName string, fieldKeys []string, fieldValue any)
	// CompareOperatorHandler is a function that generates an appropriate query condition for the given field.
	//
	// fieldName is the name of the field.
	// fieldKeys is a list of keys used only for nested fields.
	// fieldValue is the value to compare against.
	// querySettings are the settings to use for the query defining which fields are to be treated e.g. as wildcard array or keyword field.
	CompareOperatorHandler func(fieldName string, fieldKeys []string, fieldValue any,
		querySettings *QuerySettings) esquery.Mappable
)

// RatingRange represent a closed interval of float32 values.
type RatingRange struct {
	Min float32 // Lower bound of the rating range (inclusive)
	Max float32 // Upper bound of the rating range (inclusive)
}

// QuerySettings is used to configure the query builder.
type QuerySettings struct {
	// WildcardArrays is a map of field names to a boolean value indicating whether the field is to be treated as a wildcard array.
	WildcardArrays map[string]bool
	// IsEqualToKeywordFields is a map of field names to a boolean value indicating whether the field is to be treated as a keyword field.
	IsEqualToKeywordFields map[string]bool
	// UseNestedMatchQueryFields is a map of field names to a boolean value indicating whether the field is to be treated as a nested query field.
	UseNestedMatchQueryFields map[string]bool
	// NestedQueryFieldDefinitions is a list of nested query field definitions.
	NestedQueryFieldDefinitions []NestedQueryFieldDefinition
	// UseMatchPhraseFields is a map of field names to a boolean value indicating whether the field should use a match phrase query.
	UseMatchPhrase     map[string]bool
	FilterFieldMapping map[string]string
	// StringFieldRating is a map for field names with a rating. The rating is used to determine the compare order of the field in the query.
	StringFieldRating map[string]map[string]RatingRange
}

// NestedQueryFieldDefinition is a definition of a nested query field.
type NestedQueryFieldDefinition struct {
	// FieldName is the name of the field.
	FieldName string
	// FieldKeyName is the name of the key field.
	FieldKeyName string
	// FieldValueName is the name of the value field.
	FieldValueName string
}

// CompareOperator defines a mapping between a filter.CompareOperator and a function to generate an appropriate
// query condition in from of a CompareOperatorHandler.
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

// ReplaceCompareOperators replaces the set of CompareOperator to be used for this query builder.
//
// operators is the new set of CompareOperator to use.
func (q *BoolQueryBuilder) ReplaceCompareOperators(operators []CompareOperator) *BoolQueryBuilder {
	q.compareOperators = operators
	return q
}

// AddCompareOperators adds the given set of CompareOperator to the set of CompareOperator to be used for this query builder.
//
// operators is the set of CompareOperator to add.
func (q *BoolQueryBuilder) AddCompareOperators(operators ...CompareOperator) *BoolQueryBuilder {
	q.compareOperators = append(q.compareOperators, operators...)
	return q
}

// AddTermsFilter adds a terms filter to this query.
//
// values is the list of values to filter for.
func (q *BoolQueryBuilder) AddTermsFilter(fieldName string, values ...interface{}) *BoolQueryBuilder {
	q.query = q.query.Filter(esquery.Terms(fieldName, values...))
	return q
}

// AddTermFilter adds a term filter to this query.
//
// value is the value to filter for.
func (q *BoolQueryBuilder) AddTermFilter(fieldName string, value interface{}) *BoolQueryBuilder {
	q.query = q.query.Filter(esquery.Term(fieldName, value))
	return q
}

func (q *BoolQueryBuilder) addToMust(call CompareOperatorHandler) queryAppender {
	return func(fieldName string, fieldKeys []string, fieldValue any) {
		value := call(fieldName, fieldKeys, fieldValue, q.querySettings)
		if value != nil {
			q.Must = append(q.Must, value)
		}
	}
}

func (q *BoolQueryBuilder) addToMustNot(call CompareOperatorHandler) queryAppender {
	return func(fieldName string, fieldKeys []string, fieldValue any) {
		value := call(fieldName, fieldKeys, fieldValue, q.querySettings)
		if value != nil {
			q.MustNot = append(q.MustNot, value)
		}
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

	effectiveRequest, err := effectiveFilterFields(*request, q.querySettings.FilterFieldMapping)
	if err != nil {
		return err
	}

	operatorMapping := q.createOperatorMapping()

	for _, field := range effectiveRequest.Fields {
		if handler, ok := operatorMapping[field.Operator]; ok {
			handler(field.Name, field.Keys, field.Value)
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
	case "":
		return fmt.Errorf("missing mandatory field `Operator` in filter request")
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
		{Operator: filter.CompareOperatorIsIpEqualTo, Handler: HandleCompareOperatorIsKeywordEqualTo, MustCondition: true},
		{Operator: filter.CompareOperatorIsIpNotEqualTo, Handler: HandleCompareOperatorIsKeywordEqualTo, MustCondition: false},
		{Operator: filter.CompareOperatorIsStringEqualTo, Handler: HandleCompareOperatorIsKeywordEqualTo, MustCondition: true},
		{
			Operator: filter.CompareOperatorIsStringNotEqualTo,
			Handler:  HandleCompareOperatorIsKeywordEqualTo, MustCondition: false,
		},
		{Operator: filter.CompareOperatorContains, Handler: HandleCompareOperatorContains, MustCondition: true},
		{Operator: filter.CompareOperatorDoesNotContain, Handler: HandleCompareOperatorContains, MustCondition: false},
		{Operator: filter.CompareOperatorBeginsWith, Handler: HandleCompareOperatorBeginsWith, MustCondition: true},
		{
			Operator: filter.CompareOperatorDoesNotBeginWith,
			Handler:  HandleCompareOperatorNotBeginsWith, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorIsLessThanOrEqualTo,
			Handler:  HandleCompareOperatorIsLessThanOrEqualTo, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
			Handler:  HandleCompareOperatorIsGreaterThanOrEqualTo, MustCondition: true,
		},
		{Operator: filter.CompareOperatorIsGreaterThan, Handler: HandleCompareOperatorIsGreaterThan, MustCondition: true},
		{Operator: filter.CompareOperatorIsLessThan, Handler: HandleCompareOperatorIsLessThan, MustCondition: true},
		{Operator: filter.CompareOperatorAfterDate, Handler: HandleCompareOperatorIsGreaterThan, MustCondition: true},
		{Operator: filter.CompareOperatorBeforeDate, Handler: HandleCompareOperatorIsLessThan, MustCondition: true},
		{
			Operator: filter.CompareOperatorIsGreaterThanRating,
			Handler:  HandleCompareOperatorIsGreaterThanRating, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorIsLessThanRating,
			Handler:  HandleCompareOperatorIsLessThanRating, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorIsGreaterThanOrEqualToRating,
			Handler:  HandleCompareOperatorIsGreaterThanOrEqualToRating, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorIsLessThanOrEqualToRating,
			Handler:  HandleCompareOperatorIsLessThanOrEqualToRating, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorIsEqualToRating,
			Handler:  HandleCompareOperatorIsEqualToRating, MustCondition: true,
		},
		{
			Operator: filter.CompareOperatorIsNotEqualToRating,
			Handler:  HandleCompareOperatorIsNotEqualToRating, MustCondition: true,
		},
	}
}
