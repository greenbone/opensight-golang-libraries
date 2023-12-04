// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_query

import (
	"github.com/aquasecurity/esquery"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/pkg/errors"
)

type (
	QueryAppender          func(fieldName string, fieldKeys []string, fieldValue any)
	CompareOperatorHandler func(fieldName string, fieldKeys []string, fieldValue any) esquery.Mappable
)

type QuerySettings struct {
	WildcardArrays              map[string]bool
	IsEqualToKeywordFields      map[string]bool
	UseNestedMatchQueryFields   map[string]bool
	UseMatchPhrase              map[string]bool
	CompareOperators            []CompareOperator
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

var actualBoolQuerySettings QuerySettings

func SetQuerySettings(settings QuerySettings) {
	actualBoolQuerySettings = settings
}

type BoolQueryBuilder struct {
	size  uint64
	query *esquery.BoolQuery
	// aggregations aggregation of this search request.
	// Deprecated: Better create custom implementation for more verbosity.
	aggregations []Aggregation
	Must         []esquery.Mappable
	MustNot      []esquery.Mappable
}

func (q *BoolQueryBuilder) AddTermsFilter(fieldName string, values ...interface{}) *BoolQueryBuilder {
	q.query = q.query.Filter(esquery.Terms(fieldName, values...))
	return q
}

func (q *BoolQueryBuilder) AddTermFilter(fieldName string, value interface{}) *BoolQueryBuilder {
	q.query = q.query.Filter(esquery.Term(fieldName, value))
	return q
}

// AddAggregation adds an aggregation to this search request.
//
// Deprecated: Better create custom implementation for more verbosity.
func (q *BoolQueryBuilder) AddAggregation(aggregation Aggregation) *BoolQueryBuilder {
	q.aggregations = append(q.aggregations, aggregation)
	return q
}

func (q *BoolQueryBuilder) Size(size uint64) *BoolQueryBuilder {
	q.size = size
	return q
}

func (q *BoolQueryBuilder) AddToMust(call CompareOperatorHandler) QueryAppender {
	return func(fieldName string, fieldKeys []string, fieldValue any) {
		value := call(fieldName, fieldKeys, fieldValue)
		if value != nil {
			q.Must = append(q.Must, value)
		}
	}
}

func (q *BoolQueryBuilder) AddToMustNot(call CompareOperatorHandler) QueryAppender {
	return func(fieldName string, fieldKeys []string, fieldValue any) {
		value := call(fieldName, fieldKeys, fieldValue)
		if value != nil {
			q.MustNot = append(q.MustNot, value)
		}
	}
}

func (q *BoolQueryBuilder) createOperatorMapping() map[filter.CompareOperator]QueryAppender {
	operatorMapping := make(map[filter.CompareOperator]QueryAppender,
		len(actualBoolQuerySettings.CompareOperators))
	for _, setting := range actualBoolQuerySettings.CompareOperators {
		if setting.MustCondition {
			operatorMapping[setting.Operator] = q.AddToMust(setting.Handler)
		} else {
			operatorMapping[setting.Operator] = q.AddToMustNot(setting.Handler)
		}
	}
	return operatorMapping
}

func (q *BoolQueryBuilder) AddFilterRequest(request *filter.Request) error {
	if request == nil {
		return nil
	}

	effectiveRequest, err := EffectiveFilterFields(*request)
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

func (q *BoolQueryBuilder) Build() *esquery.BoolQuery {
	return q.query
}

// ToJson returns a json representation of the search request
//
// Deprecated: do not use due to dubious size setting. Better create custom implementation.
func (q *BoolQueryBuilder) ToJson() (json string, err error) {
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

func NewBoolQueryBuilder() *BoolQueryBuilder {
	return &BoolQueryBuilder{
		query: esquery.Bool(),
	}
}

func NewBoolQueryBuilderWith(query *esquery.BoolQuery) *BoolQueryBuilder {
	return &BoolQueryBuilder{
		query: query,
	}
}
