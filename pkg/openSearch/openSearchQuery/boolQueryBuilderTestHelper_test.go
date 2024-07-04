// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"github.com/aquasecurity/esquery"
)

// Aggregation is an interface for all aggregations that can be converted to an esquery.Aggregation.
type Aggregation interface {
	ToEsAggregation() esquery.Aggregation
}

// testBoolQueryBuilderWrapper is a wrapper around a BoolQueryBuilder that provides additional methods for testing.
type testBoolQueryBuilderWrapper struct {
	*BoolQueryBuilder
	// aggregations aggregation of this search request.
	// Only used for testing. Create custom implementation for more verbosity.
	aggregations []Aggregation
}

// addAggregation adds an aggregation to this query.
func (q *testBoolQueryBuilderWrapper) addAggregation(aggregation Aggregation) *testBoolQueryBuilderWrapper {
	q.aggregations = append(q.aggregations, aggregation)
	return q
}

// toJson returns a json representation of the search request
//
// Only used for testing. Do not use in production code due to dubious size setting. Better create custom implementation.
func (q *testBoolQueryBuilderWrapper) toJson() (json string, err error) {
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
		return "", err1
	}
	return string(jsonByte), nil
}
