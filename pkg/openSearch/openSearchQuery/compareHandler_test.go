// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"testing"
	"time"

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
			"RatingIsGreaterThan",
			HandleCompareOperatorIsGreaterThanRating,
			"severityClass",
			nil,
			"Medium",
			esquery.Range("severityClass").Gt(float32(6.9)),
		},
		{
			"RatingIsLowerThan",
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

func TestHandleCompareOperatorDateRange(t *testing.T) {
	startDate := time.Date(2023, 2, 27, 12, 34, 56, 0, time.UTC)
	endDate := time.Date(2024, 8, 24, 0, 0, 0, 0, time.UTC) // Adjusted to match test case

	// Correct string representations in RFC3339 format
	startDateStr := "2023-02-27T12:34:56Z"
	endDateStr := "2024-08-24T00:00:00Z"

	// Parse expected values from strings
	parsedStartDate, _ := time.Parse(time.RFC3339, startDateStr)
	parsedEndDate, _ := time.Parse(time.RFC3339, endDateStr)

	field := "event.timestamp" // same for all tests
	tests := []struct {
		name     string
		field    string
		keys     []string
		value    any
		expected esquery.Mappable
	}{
		{
			name: "ValidDateRange_StringInput",
			value: []string{
				startDateStr,
				endDateStr,
			},
			expected: esquery.Range(field).
				Gte(parsedStartDate).
				Lte(parsedEndDate),
		},
		{
			name: "ValidDateRange_TimeInput",
			value: []time.Time{
				startDate,
				endDate,
			},
			expected: esquery.Range(field).
				Gte(startDate).
				Lte(endDate),
		},
		{
			name:     "InvalidDateString",
			value:    []string{"invalid-date", "2024-08-24T00:00:00Z"},
			expected: esquery.MatchNone(),
		},
		{
			name:     "NonTimeValue",
			value:    12345,
			expected: esquery.MatchNone(),
		},
		{
			name:     "InvalidSliceLength",
			value:    []string{"2023-02-27T12:34:56Z"}, // Only one date instead of two
			expected: esquery.MatchNone(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// `fieldKeys` and `QuerySettings` are irrelevant for handler
			assert.Equal(t, tt.expected, HandleCompareOperatorBetweenDates(field, nil, tt.value, nil))
		})
	}
}
func TestHandleCompareOperatorTextContains(t *testing.T) {
	field := "description" // same for all tests
	tests := []struct {
		name     string
		field    string
		keys     []string
		value    any
		expected esquery.Mappable
	}{
		{
			name:     "SingleValue",
			value:    "test",
			expected: esquery.Match(field, "test").MinimumShouldMatch("100%"),
		},
		{
			name:     "MultipleValues",
			value:    []any{"test", "example"},
			expected: esquery.Match(field, "test example").MinimumShouldMatch("100%"),
		},
		{
			name:  "EmptyValue",
			value: "",
			// handled gracefully by openSearch, will give no results
			expected: esquery.Match(field, "").MinimumShouldMatch("100%"),
		},
		{
			name:  "EmptySlice",
			value: []any{},
			// handled gracefully by openSearch, will give no results
			expected: esquery.Match(field, "").MinimumShouldMatch("100%"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// `fieldKeys` and `QuerySettings` are irrelevant for handler
			assert.Equal(t, tt.expected, HandleCompareOperatorTextContains(field, nil, tt.value, nil))
		})
	}
}

func TestHandleCompareOperatorBeginsWith(t *testing.T) {
	field := "testField"

	tests := []struct {
		name     string
		value    any
		expected esquery.Mappable
	}{
		{
			name:     "SingleValue",
			value:    "test",
			expected: esquery.Prefix(field, "test"),
		},
		{
			name:  "MultipleValues",
			value: []any{"test1", "test2"},
			expected: esquery.Bool().
				Should(
					esquery.Prefix(field, "test1"),
					esquery.Prefix(field, "test2"),
				).
				MinimumShouldMatch(1),
		},
		{
			name:     "EmptyValue",
			value:    "",
			expected: esquery.Prefix(field, ""),
		},
		{
			name:  "EmptySlice",
			value: []any{},
			expected: esquery.Bool().
				Should().
				MinimumShouldMatch(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, handleCompareOperatorBeginsWith(field, tt.value))
		})
	}
}
