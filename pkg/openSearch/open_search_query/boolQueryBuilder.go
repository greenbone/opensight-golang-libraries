// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_query

import (
	"github.com/aquasecurity/esquery"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/pkg/errors"
)

type boolQueryBuilder struct {
	querySettings    *QuerySettings
	compareOperators []CompareOperator
	size             uint64
	query            *esquery.BoolQuery
	// aggregations aggregation of this search request.
	// Deprecated: Better create custom implementation for more verbosity.
	aggregations []Aggregation
	Must         []esquery.Mappable
	MustNot      []esquery.Mappable
}

type (
	QueryAppender          func(fieldName string, fieldKeys []string, fieldValue any)
	CompareOperatorHandler func(fieldName string, fieldKeys []string, fieldValue any,
		querySettings *QuerySettings) esquery.Mappable
)

type QuerySettings struct {
	WildcardArrays              map[string]bool
	IsEqualToKeywordFields      map[string]bool
	UseNestedMatchQueryFields   map[string]bool
	UseMatchPhrase              map[string]bool
	NestedQueryFieldDefinitions []NestedQueryFieldDefinition
	FilterFieldMapping          map[string]string
}

type NestedQueryFieldDefinition struct {
	FieldName      string
	FieldKeyName   string
	FieldValueName string
}

type CompareOperator struct {
	Operator      filter.CompareOperator
	Handler       CompareOperatorHandler
	MustCondition bool
}

func NewBoolQueryBuilder(querySettings *QuerySettings) *boolQueryBuilder {
	return NewBoolQueryBuilderWith(esquery.Bool(), querySettings)
}

func NewBoolQueryBuilderWith(query *esquery.BoolQuery, querySettings *QuerySettings) *boolQueryBuilder {
	return &boolQueryBuilder{
		querySettings:    querySettings,
		compareOperators: defaultCompareOperators(),
		query:            query,
	}
}

func (q *boolQueryBuilder) ReplaceCompareOperators(operators []CompareOperator) *boolQueryBuilder {
	q.compareOperators = operators
	return q
}

func (q *boolQueryBuilder) AddCompareOperators(operators ...CompareOperator) *boolQueryBuilder {
	q.compareOperators = append(q.compareOperators, operators...)
	return q
}

func (q *boolQueryBuilder) AddTermsFilter(fieldName string, values ...interface{}) *boolQueryBuilder {
	q.query = q.query.Filter(esquery.Terms(fieldName, values...))
	return q
}

func (q *boolQueryBuilder) AddTermFilter(fieldName string, value interface{}) *boolQueryBuilder {
	q.query = q.query.Filter(esquery.Term(fieldName, value))
	return q
}

// AddAggregation adds an aggregation to this search request.
//
// Deprecated: Better create custom implementation for more verbosity.
func (q *boolQueryBuilder) AddAggregation(aggregation Aggregation) *boolQueryBuilder {
	q.aggregations = append(q.aggregations, aggregation)
	return q
}

func (q *boolQueryBuilder) Size(size uint64) *boolQueryBuilder {
	q.size = size
	return q
}

func (q *boolQueryBuilder) AddToMust(call CompareOperatorHandler) QueryAppender {
	return func(fieldName string, fieldKeys []string, fieldValue any) {
		value := call(fieldName, fieldKeys, fieldValue, q.querySettings)
		if value != nil {
			q.Must = append(q.Must, value)
		}
	}
}

func (q *boolQueryBuilder) AddToMustNot(call CompareOperatorHandler) QueryAppender {
	return func(fieldName string, fieldKeys []string, fieldValue any) {
		value := call(fieldName, fieldKeys, fieldValue, q.querySettings)
		if value != nil {
			q.MustNot = append(q.MustNot, value)
		}
	}
}

func (q *boolQueryBuilder) createOperatorMapping() map[filter.CompareOperator]QueryAppender {
	operatorMapping := make(map[filter.CompareOperator]QueryAppender,
		len(q.compareOperators))
	for _, setting := range q.compareOperators {
		if setting.MustCondition {
			operatorMapping[setting.Operator] = q.AddToMust(setting.Handler)
		} else {
			operatorMapping[setting.Operator] = q.AddToMustNot(setting.Handler)
		}
	}
	return operatorMapping
}

func (q *boolQueryBuilder) AddFilterRequest(request *filter.Request) error {
	if request == nil {
		return nil
	}

	effectiveRequest, err := EffectiveFilterFields(*request, q.querySettings.FilterFieldMapping)
	if err != nil {
		return errors.WithStack(err)
	}

	operatorMapping := q.createOperatorMapping()

	for _, field := range effectiveRequest.Fields {
		if handler, ok := operatorMapping[field.Operator]; ok {
			handler(field.Name, field.Keys, field.Value)
		} else {
			return errors.Errorf("field '%s' with unknown operator '%s'", field.Name, field.Operator)
		}
	}
	switch effectiveRequest.Operator {
	case "": // empty filter request
		return nil
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
		return errors.Errorf("unknown operator '%s'", effectiveRequest.Operator)
	}
}

func (q *boolQueryBuilder) Build() *esquery.BoolQuery {
	return q.query
}

// ToJson returns a json representation of the search request
//
// Deprecated: do not use due to dubious size setting. Better create custom implementation.
func (q *boolQueryBuilder) ToJson() (json string, err error) {
	size := q.size
	if size == 0 {
		// TODO: 15.08.2022 stolksdorf - current default size is 100 until we get paging
		size = 100
	}

	var jsonByte []byte
	var err1 error

	if q.aggregations == nil {
		jsonByte, err1 = esquery.Search().
			Query(q.query).
			Size(size).
			MarshalJSON()
	} else {
		var aggregations []esquery.Aggregation
		for _, aggregation := range q.aggregations {
			aggregations = append(aggregations, aggregation.ToEsAggregation())
		}
		jsonByte, err1 = esquery.Search().
			Query(q.query).
			Aggs(aggregations...).
			Size(size).
			MarshalJSON()
	}

	if err1 != nil {
		return "", errors.WithStack(err1)
	}
	return string(jsonByte), nil
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
	}
}
