// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/testFolder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO free from JSON generation for easier testing?
func TestBoolQueryBuilder(t *testing.T) {
	var (
		query  *BoolQueryBuilder
		folder testFolder.TestFolder
	)

	setup := func(t *testing.T) {
		querySettings := QuerySettings{
			FilterFieldMapping: map[string]string{"testName": "testName"},
		}
		query = NewBoolQueryBuilder(&querySettings)
		folder = testFolder.NewTestFolder()
	}

	t.Run("shouldReturnJsonForFilterTerm", func(t *testing.T) {
		setup(t)

		query.AddTermFilter("foo", "bar")

		json, err := query.toJson()
		require.NoError(t, err)
		assert.JSONEq(t, folder.GetContent(t, "testdata/filterTerm.json"), json)
	})

	t.Run("should work with empty filter request", func(t *testing.T) {
		setup(t)

		err := query.AddFilterRequest(nil)

		require.NoError(t, err)
	})

	t.Run("shouldReturnJsonForFilterTermWithFilterRequest", func(t *testing.T) {
		setup(t)

		query.AddTermFilter("foo", "bar")
		err := query.AddFilterRequest(&filter.Request{
			Operator: filter.LogicOperatorAnd,
			Fields: []filter.RequestField{
				{
					Name:     "testName",
					Operator: filter.CompareOperatorBeginsWith,
					Value:    "start",
				},
			},
		})
		require.NoError(t, err)

		json, err := query.toJson()
		require.NoError(t, err)
		assert.JSONEq(t, folder.GetContent(t, "testdata/filterTermWithFilterRequest.json"), json)
	})
}

func TestFilterQueryOperatorAnd(t *testing.T) {
	querySettings := QuerySettings{
		FilterFieldMapping: map[string]string{"testName": "testName"},
	}

	mixedTests := map[string]struct {
		file     string
		operator filter.CompareOperator
		value    any
	}{
		"shouldReturnJsonForOperatorBeginsWithSingleValue": {
			file:     "testdata/And/singleValue/BeginsWith.json",
			operator: filter.CompareOperatorBeginsWith, value: 5,
		},
		"shouldReturnJsonForOperatorDoesNotBeginsWithSingleValue": {
			file:     "testdata/And/singleValue/DoesNotBeginsWith.json",
			operator: filter.CompareOperatorDoesNotBeginWith, value: "5",
		},
		"shouldReturnJsonForOperatorDoesNotBeginsWithMultiValue": {
			file:     "testdata/And/multiValue/DoesNotBeginsWith.json",
			operator: filter.CompareOperatorDoesNotBeginWith, value: []interface{}{"5", "6"},
		},
		"shouldReturnJsonForOperatorBeginsWithMultiValue": {
			file: "testdata/And/multiValue/BeginsWith.json", operator: filter.CompareOperatorBeginsWith, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorContainsSingleValue": {
			file: "testdata/And/singleValue/Contains.json", operator: filter.CompareOperatorContains, value: "test",
		},
		"shouldReturnJsonForOperatorContainsMultiValue": {
			file: "testdata/And/multiValue/Contains.json", operator: filter.CompareOperatorContains,
			value: []interface{}{"test1", "test2"},
		},
		"shouldReturnJsonForOperatorDoesNotContainSingleValue": {
			file:     "testdata/And/singleValue/DoesNotContain.json",
			operator: filter.CompareOperatorDoesNotContain, value: "test",
		},
		"shouldReturnJsonForOperatorDoesNotContainMultiValue": {
			file:     "testdata/And/multiValue/DoesNotContain.json",
			operator: filter.CompareOperatorDoesNotContain,
			value:    []interface{}{"test1", "test2"},
		},
		"shouldReturnJsonForOperatorIsStringEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsStringEqualTo.json",
			operator: filter.CompareOperatorIsStringEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsStringEqualToMultiValue": {
			file:     "testdata/And/multiValue/IsStringEqualTo.json",
			operator: filter.CompareOperatorIsStringEqualTo, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorIsStringNotEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsStringNotEqualTo.json",
			operator: filter.CompareOperatorIsStringNotEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsStringNotEqualToMultiValue": {
			file:     "testdata/And/multiValue/IsStringNotEqualTo.json",
			operator: filter.CompareOperatorIsStringNotEqualTo, value: []interface{}{5, 6},
		},
	}
	for name, tc := range mixedTests {
		t.Run(name, func(t *testing.T) {
			query := NewBoolQueryBuilder(&querySettings)
			err := query.AddFilterRequest(&filter.Request{
				Operator: filter.LogicOperatorAnd,
				Fields: []filter.RequestField{
					{
						Name:     "testName",
						Operator: tc.operator,
						Value:    tc.value,
					},
				},
			})
			require.NoError(t, err)
			json, err := query.toJson()
			require.NoError(t, err)

			expectedJson := testFolder.NewTestFolder().
				GetContent(t, tc.file)
			assert.JSONEq(t, expectedJson, json)
		})
	}

	singleValueTests := map[string]struct {
		file     string
		operator filter.CompareOperator
	}{
		"shouldReturnJsonForOperatorIsGreaterThan": {
			file:     "testdata/And/singleValue/IsGreaterThan.json",
			operator: filter.CompareOperatorIsGreaterThan,
		},
		"shouldReturnJsonForOperatorIsGreaterThanOrEqualTo": {
			file:     "testdata/And/singleValue/IsGreaterThanOrEqualTo.json",
			operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
		},
		"shouldReturnJsonForOperatorIsLessThan": {
			file:     "testdata/And/singleValue/IsLessThan.json",
			operator: filter.CompareOperatorIsLessThan,
		},
		"shouldReturnJsonForOperatorIsLessThanOrEqualTo": {
			file:     "testdata/And/singleValue/IsLessThanOrEqualTo.json",
			operator: filter.CompareOperatorIsLessThanOrEqualTo,
		},
	}
	for name, tc := range singleValueTests {
		t.Run(name, func(t *testing.T) {
			query := NewBoolQueryBuilder(&querySettings)
			err := query.AddFilterRequest(&filter.Request{
				Operator: filter.LogicOperatorAnd,
				Fields: []filter.RequestField{
					{
						Name:     "testName",
						Operator: tc.operator,
						Value:    5,
					},
				},
			})
			require.NoError(t, err)
			json, err := query.toJson()
			require.NoError(t, err)

			assert.JSONEq(t, testFolder.NewTestFolder().
				GetContent(t, tc.file), json)
		})
	}
}

func TestFilterQueryOperatorOr(t *testing.T) {
	querySettings := QuerySettings{
		FilterFieldMapping: map[string]string{"testName": "testName"},
	}

	mixedTests := map[string]struct {
		file     string
		operator filter.CompareOperator
		value    any
	}{
		"shouldReturnJsonForOperatorBeginsWithSingleValue": {
			file: "testdata/Or/singleValue/BeginsWith.json", operator: filter.CompareOperatorBeginsWith, value: 5,
		},
		"shouldReturnJsonForOperatorBeginsWithMultiValue": {
			file: "testdata/Or/multiValue/BeginsWith.json", operator: filter.CompareOperatorBeginsWith, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorContainsSingleValue": {
			file: "testdata/Or/singleValue/Contains.json", operator: filter.CompareOperatorContains, value: "test",
		},
		"shouldReturnJsonForOperatorContainsMultiValue": {
			file: "testdata/Or/multiValue/Contains.json", operator: filter.CompareOperatorContains,
			value: []interface{}{"test1", "test2"},
		},
		"shouldReturnJsonForOperatorDoesNotContainSingleValue": {
			file:     "testdata/Or/singleValue/DoesNotContain.json",
			operator: filter.CompareOperatorDoesNotContain, value: "test1",
		},
		"shouldReturnJsonForOperatorDoesNotContainMultiValue": {
			file:     "testdata/Or/multiValue/DoesNotContain.json",
			operator: filter.CompareOperatorDoesNotContain,
			value:    []interface{}{"test1", "test2"},
		},
		"shouldReturnJsonForOperatorIsStringEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsStringEqualTo.json",
			operator: filter.CompareOperatorIsStringEqualTo, value: "test",
		},
		"shouldReturnJsonForOperatorIsStringEqualToMultiValue": {
			file:     "testdata/Or/multiValue/IsStringEqualTo.json",
			operator: filter.CompareOperatorIsStringEqualTo, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorIsStringNotEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsStringNotEqualTo.json",
			operator: filter.CompareOperatorIsStringNotEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsStringNotEqualToMultiValue": {
			file:     "testdata/Or/multiValue/IsStringNotEqualTo.json",
			operator: filter.CompareOperatorIsStringNotEqualTo, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorIsEqualToSingleValue": {
			file: "testdata/Or/singleValue/IsEqualTo.json", operator: filter.CompareOperatorIsEqualTo, value: "test",
		},
		"shouldReturnJsonForOperatorIsEqualToMultiValue": {
			file: "testdata/Or/multiValue/IsEqualTo.json", operator: filter.CompareOperatorIsEqualTo, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorIsNotEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsNotEqualTo.json",
			operator: filter.CompareOperatorIsNotEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsNotEqualToMultiValue": {
			file:     "testdata/Or/multiValue/IsNotEqualTo.json",
			operator: filter.CompareOperatorIsNotEqualTo, value: []interface{}{5, 6},
		},
	}
	for name, tc := range mixedTests {
		t.Run(name, func(t *testing.T) {
			query := NewBoolQueryBuilder(&querySettings)
			err := query.AddFilterRequest(&filter.Request{
				Operator: filter.LogicOperatorOr,
				Fields: []filter.RequestField{
					{
						Name:     "testName",
						Operator: tc.operator,
						Value:    tc.value,
					},
				},
			})
			require.NoError(t, err)
			json, err := query.toJson()
			require.NoError(t, err)

			assert.JSONEq(t, testFolder.NewTestFolder().
				GetContent(t, tc.file), json)
		})
	}

	singleValueTests := map[string]struct {
		file     string
		operator filter.CompareOperator
	}{
		"shouldReturnJsonForOperatorIsGreaterThan": {
			file:     "testdata/And/singleValue/IsGreaterThan.json",
			operator: filter.CompareOperatorIsGreaterThan,
		},
		"shouldReturnJsonForOperatorIsGreaterThanOrEqualTo": {
			file:     "testdata/And/singleValue/IsGreaterThanOrEqualTo.json",
			operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
		},
		"shouldReturnJsonForOperatorIsLessThan": {
			file:     "testdata/And/singleValue/IsLessThan.json",
			operator: filter.CompareOperatorIsLessThan,
		},
		"shouldReturnJsonForOperatorIsLessThanOrEqualTo": {
			file:     "testdata/And/singleValue/IsLessThanOrEqualTo.json",
			operator: filter.CompareOperatorIsLessThanOrEqualTo,
		},
	}
	for name, tc := range singleValueTests {
		t.Run(name, func(t *testing.T) {
			query := NewBoolQueryBuilder(&querySettings)
			err := query.AddFilterRequest(&filter.Request{
				Operator: filter.LogicOperatorAnd,
				Fields: []filter.RequestField{
					{
						Name:     "testName",
						Operator: tc.operator,
						Value:    5,
					},
				},
			})
			require.NoError(t, err)
			json, err := query.toJson()
			require.NoError(t, err)

			assert.JSONEq(t, testFolder.NewTestFolder().
				GetContent(t, tc.file), json)
		})
	}
}
