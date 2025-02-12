// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"testing"

	"github.com/aquasecurity/esquery"
	esextensions "github.com/greenbone/opensight-golang-libraries/pkg/openSearch/esextension"
	"github.com/stretchr/testify/assert"
)

func TestHandleCompareOperator(t *testing.T) {
	querySettings := QuerySettings{
		WildcardArrays: map[string]bool{
			"asset.ips":          true,
			"asset.macAddresses": true,
		},
		UseMatchPhrase: map[string]bool{
			"vulnerabilityTest.family": true,
		},
		NestedQueryFieldDefinitions: []NestedQueryFieldDefinition{
			{
				FieldName:      "asset.tags",
				FieldKeyName:   "asset.tags.tagname",
				FieldValueName: "asset.tags.tagvalue",
			},
		},
	}

	tests := []struct {
		name     string
		handler  CompareOperatorHandler
		field    string
		keys     []string
		value    any
		expected esquery.Mappable
	}{
		{
			"IsEqualTo",
			HandleCompareOperatorIsEqualTo,
			"asset.ips",
			nil,
			"10.0.0.1",
			esquery.Term("asset.ips", "10.0.0.1"),
		},
		{
			"IsKeywordEqualTo",
			HandleCompareOperatorIsKeywordEqualTo,
			"asset.hostnames",
			nil,
			"example.com",
			esquery.Term("asset.hostnames.keyword", "example.com"),
		},
		{
			"Contains",
			HandleCompareOperatorContains,
			"asset.macAddresses",
			nil,
			"00:1A:2B:3C:4D:5E",
			esquery.Wildcard("asset.macAddresses", "*00:1A:2B:3C:4D:5E*"),
		},
		{
			"nestedHandleCompareOperatorContains",
			nestedHandleCompareOperatorContains,
			"asset.tags",
			[]string{"name"},
			"value",
			esextensions.Nested("asset.tags.tagname", *esquery.Bool().
				Must(
					esquery.Match("asset.tags.tagname", "name"),
					esquery.Wildcard("asset.tags.tagvalue", "*value*"),
				),
			),
		},
		{
			"simpleNestedMatchQuery",
			simpleNestedMatchQuery,
			"asset.tags",
			[]string{"name"},
			"value",
			esextensions.Nested("asset.tags.tagname", *esquery.Bool().
				Must(
					esquery.Match("asset.tags.tagname", "name"),
					esquery.Match("asset.tags.tagvalue", "value"),
				),
			),
		},
		{
			name:     "matchPhraseFieldQuery",
			handler:  HandleCompareOperatorIsEqualTo,
			field:    "vulnerabilityTest.family",
			keys:     nil,
			value:    "denial of service",
			expected: esquery.MatchPhrase("vulnerabilityTest.family", "denial of service"),
		},
		// Add other test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.handler(tt.field, tt.keys, tt.value, &querySettings))
		})
	}
}

func TestHandleRatingComparison(t *testing.T) {
	querySettings := QuerySettings{
		StringFieldRating: map[string]map[string]RatingRange{
			"severityClass": {
				"Log":      {0, 0},
				"Low":      {0.1, 3.9},
				"Medium":   {4, 6.9},
				"High":     {7, 8.9},
				"Critical": {9, 10},
			},
		},
	}

	tests := []struct {
		name     string
		handler  CompareOperatorHandler
		field    string
		keys     []string
		value    any
		expected esquery.Mappable
	}{
		{
			"RatingIsGreaterThen",
			HandleCompareOperatorIsGreaterThanRating,
			"severityClass",
			nil,
			"Medium",
			esquery.Range("severityClass").Gt(float32(6.9)),
		},
		{
			"FailingRatingIsLowerThen",
			HandleCompareOperatorIsLessThanRating,
			"severityClass",
			nil,
			"Medium",
			esquery.Range("severityClass").Lt(float32(4)),
		},
		{
			"GreaterOrEqualTo",
			HandleCompareOperatorIsGreaterThanOrEqualToRating,
			"severityClass",
			nil,
			"High",
			esquery.Range("severityClass").Gte(float32(7)),
		},
		{
			"LessOrEqualTo",
			HandleCompareOperatorIsLessThanOrEqualToRating,
			"severityClass",
			nil,
			"High",
			esquery.Range("severityClass").Lte(float32(8.9)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.handler(tt.field, tt.keys, tt.value, &querySettings))
		})
	}
}
