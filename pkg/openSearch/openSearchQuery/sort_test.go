// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"testing"

	"github.com/aquasecurity/esquery"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
	"github.com/stretchr/testify/assert"
)

type SortingTestCase struct {
	SortingRequest       *sorting.Request
	ExpectedQueryJson    string
	ExpectedErrorMessage string
}

func TestSorting(t *testing.T) {
	testCases := map[string]SortingTestCase{
		"no sorting": {
			SortingRequest:    nil,
			ExpectedQueryJson: `{"aggs":{"vulnerabilityWithAssetCountAgg":{"terms":{"field":"vulnerabilityTest.oid.keyword"}}},"query":{"bool":{}},"size":0}`,
		},
		"sorting by severity": {
			SortingRequest: &sorting.Request{
				SortColumn:    "severity",
				SortDirection: "desc",
			},
			ExpectedQueryJson: `{"aggs":{"vulnerabilityWithAssetCountAgg":{"aggs":{"maxSeverity":{"max":{"field":"vulnerabilityTest.severityCvss.override"}}},"terms":{"field":"vulnerabilityTest.oid.keyword","order":{"maxSeverity.value":"DESC"}}}},"query":{"bool":{}},"size":0}`,
		},
		"sorting by qod": {
			SortingRequest: &sorting.Request{
				SortColumn:    "qod",
				SortDirection: "desc",
			},
			ExpectedQueryJson: `{"aggs":{"vulnerabilityWithAssetCountAgg":{"aggs":{"maxQod":{"max":{"field":"qod"}}},"terms":{"field":"vulnerabilityTest.oid.keyword","order":{"maxQod.value":"DESC"}}}},"query":{"bool":{}},"size":0}`,
		},
		"sorting by unknown Field": {
			SortingRequest: &sorting.Request{
				SortColumn:    "unknown",
				SortDirection: "desc",
			},
			ExpectedQueryJson:    `{"aggs":{"vulnerabilityWithAssetCountAgg":{"terms":{"field":"vulnerabilityTest.oid.keyword"}}},"query":{"bool":{}},"size":0}`,
			ExpectedErrorMessage: "unknown is no valid sort column, possible values:",
		},
		"sorting qod asc": {
			SortingRequest: &sorting.Request{
				SortColumn:    "qod",
				SortDirection: "asc",
			},
			ExpectedQueryJson:    `{"aggs":{"vulnerabilityWithAssetCountAgg":{"aggs":{"maxQod":{"max":{"field":"qod"}}},"terms":{"field":"vulnerabilityTest.oid.keyword","order":{"maxQod.value":"ASC"}}}},"query":{"bool":{}},"size":0}`,
			ExpectedErrorMessage: "",
		},
	}

	for name := range testCases {
		t.Run(name, func(t *testing.T) {
			q := NewBoolQueryBuilder(&QuerySettings{})
			subAggs := []esquery.Aggregation{}
			subAggs, err := AddMaxAggForSorting(subAggs, testCases[name].SortingRequest, sortFieldMapping)
			if testCases[name].ExpectedErrorMessage != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCases[name].ExpectedErrorMessage)
				return
			} else {
				assert.Nil(t, err)
			}
			termsAggregation := esquery.TermsAgg("vulnerabilityWithAssetCountAgg", "vulnerabilityTest.oid.keyword").
				Aggs(subAggs...)
			termsAggregation, err = AddOrder(termsAggregation,
				testCases[name].SortingRequest, sortFieldMapping)
			resultingJson, queryErr := esquery.Search().Query(q.Build()).Aggs(termsAggregation).Size(0).MarshalJSON()
			assert.Nil(t, queryErr)
			assert.JSONEq(t, testCases[name].ExpectedQueryJson, string(resultingJson))
		})
	}
}

var sortFieldMapping = map[string]EffectiveSortField{
	"severity": {
		PlainField:       strPtr("vulnerabilityTest.severityCvss.override"),
		AggregationName:  strPtr("maxSeverity"),
		AggregationValue: "maxSeverity.value",
	},
	"qod": {
		PlainField:       strPtr("qod"),
		AggregationName:  strPtr("maxQod"),
		AggregationValue: "maxQod.value",
	},
}

func strPtr(s string) *string {
	return &s
}
