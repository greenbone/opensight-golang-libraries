// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
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

var querySettings = QuerySettings{
	FilterFieldMapping: map[string]string{"testName": "testName"},
}

var emptyBoolQueryJSON = `{"query":{"bool":{}}}`

func TestBoolQueryBuilder_BasicFunctionality(t *testing.T) {
	tests := map[string]struct {
		filterRequest *filter.Request
		wantJSON      string
		wantErr       bool
	}{
		"should work with empty filter request": {
			filterRequest: nil,
			wantJSON:      emptyBoolQueryJSON,
			wantErr:       false,
		},
		"should work with empty (non-nil) filter request": {
			filterRequest: &filter.Request{},
			wantJSON:      emptyBoolQueryJSON,
			wantErr:       false,
		},
		"should fail with invalid filter request (missing logic operator)": {
			filterRequest: &filter.Request{
				Fields: []filter.RequestField{
					{
						Name:     "testName",
						Operator: filter.CompareOperatorBeginsWith,
						Value:    "start",
					},
				},
				// missing `Operator`
			},
			wantErr: true,
		},
		"should fail with invalid filter request (empty field name)": {
			filterRequest: &filter.Request{
				Fields: []filter.RequestField{
					{
						Name:     "",
						Operator: filter.CompareOperatorBeginsWith,
						Value:    "start",
					},
				},
				Operator: filter.LogicOperatorAnd,
			},
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			query := testBoolQueryBuilderWrapper{}
			query.BoolQueryBuilder = NewBoolQueryBuilder(&querySettings)
			err := query.AddFilterRequest(tc.filterRequest)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				json, err := query.toJson()
				require.NoError(t, err)
				assert.JSONEq(t, tc.wantJSON, json)
			}
		})
	}
}

func TestBoolQueryBuilder_TermFilter(t *testing.T) {
	var (
		query  testBoolQueryBuilderWrapper
		folder testFolder.TestFolder
	)

	setup := func() {
		query = testBoolQueryBuilderWrapper{}
		query.BoolQueryBuilder = NewBoolQueryBuilder(&querySettings)
		folder = testFolder.NewTestFolder()
	}

	t.Run("shouldReturnJsonForFilterTerm", func(t *testing.T) {
		setup()

		query.AddTermFilter("foo", "bar")

		json, err := query.toJson()
		require.NoError(t, err)
		assert.JSONEq(t, folder.GetContent(t, "testdata/filterTerm.json"), json)
	})

	t.Run("shouldReturnJsonForFilterTermWithFilterRequest", func(t *testing.T) {
		setup()

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
	mixedTests := map[string]struct {
		file     string
		operator filter.CompareOperator
		value    any
	}{
		// multi value
		"shouldReturnJsonForOperatorBeginsWithMultiValue": {
			file: "testdata/And/multiValue/BeginsWith.json", operator: filter.CompareOperatorBeginsWith, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorContainsMultiValue": {
			file: "testdata/And/multiValue/Contains.json", operator: filter.CompareOperatorContains,
			value: []interface{}{"test1", "test2"},
		},
		"shouldReturnJsonForOperatorDoesNotBeginsWithMultiValue": {
			file:     "testdata/And/multiValue/DoesNotBeginsWith.json",
			operator: filter.CompareOperatorDoesNotBeginWith, value: []interface{}{"5", "6"},
		},
		"shouldReturnJsonForOperatorDoesNotContainMultiValue": {
			file:     "testdata/And/multiValue/DoesNotContain.json",
			operator: filter.CompareOperatorDoesNotContain,
			value:    []interface{}{"test1", "test2"},
		},
		"shouldReturnJsonForOperatorIsStringEqualToMultiValue": {
			file:     "testdata/And/multiValue/IsStringEqualTo.json",
			operator: filter.CompareOperatorIsStringEqualTo, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorIsStringNotEqualToMultiValue": {
			file:     "testdata/And/multiValue/IsStringNotEqualTo.json",
			operator: filter.CompareOperatorIsStringNotEqualTo, value: []interface{}{5, 6},
		},
		// single value
		"shouldReturnJsonForOperatorAfterDate": {
			file:     "testdata/And/singleValue/AfterDate.json",
			operator: filter.CompareOperatorAfterDate, value: "2024-08-23T00:00:00.000+02:00",
		},
		"shouldReturnJsonForOperatorBeforeDate": {
			file:     "testdata/And/singleValue/BeforeDate.json",
			operator: filter.CompareOperatorBeforeDate, value: "2024-08-23T00:00:00.000+02:00",
		},
		"shouldReturnJsonForOperatorBeginsWithSingleValue": {
			file:     "testdata/And/singleValue/BeginsWith.json",
			operator: filter.CompareOperatorBeginsWith, value: 5,
		},
		"shouldReturnJsonForOperatorContainsSingleValue": {
			file: "testdata/And/singleValue/Contains.json", operator: filter.CompareOperatorContains, value: "test",
		},
		"shouldReturnJsonForOperatorDoesNotBeginsWithSingleValue": {
			file:     "testdata/And/singleValue/DoesNotBeginWith.json",
			operator: filter.CompareOperatorDoesNotBeginWith, value: "5",
		},
		"shouldReturnJsonForOperatorDoesNotContainSingleValue": {
			file:     "testdata/And/singleValue/DoesNotContain.json",
			operator: filter.CompareOperatorDoesNotContain, value: "test",
		},
		"shouldReturnJsonForOperatorIsEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsEqualTo.json",
			operator: filter.CompareOperatorIsEqualTo, value: "5",
		},
		"shouldReturnJsonForOperatorIsGreaterThan": {
			file:     "testdata/And/singleValue/IsGreaterThan.json",
			operator: filter.CompareOperatorIsGreaterThan, value: 5,
		},
		"shouldReturnJsonForOperatorIsGreaterThanOrEqualTo": {
			file:     "testdata/And/singleValue/IsGreaterThanOrEqualTo.json",
			operator: filter.CompareOperatorIsGreaterThanOrEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsIpEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsIpEqualTo.json",
			operator: filter.CompareOperatorIsIpEqualTo, value: "127.0.0.1",
		},
		"shouldReturnJsonForOperatorIsIpNotEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsIpNotEqualTo.json",
			operator: filter.CompareOperatorIsIpNotEqualTo, value: "127.0.0.1",
		},
		"shouldReturnJsonForOperatorIsLessThan": {
			file:     "testdata/And/singleValue/IsLessThan.json",
			operator: filter.CompareOperatorIsLessThan, value: 5,
		},
		"shouldReturnJsonForOperatorIsLessThanOrEqualTo": {
			file:     "testdata/And/singleValue/IsLessThanOrEqualTo.json",
			operator: filter.CompareOperatorIsLessThanOrEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsNotEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsNotEqualTo.json",
			operator: filter.CompareOperatorIsNotEqualTo, value: "5",
		},
		"shouldReturnJsonForOperatorIsNumberEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsNumberEqualTo.json",
			operator: filter.CompareOperatorIsNumberEqualTo, value: "5",
		},
		"shouldReturnJsonForOperatorIsNumberNotEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsNumberNotEqualTo.json",
			operator: filter.CompareOperatorIsNumberNotEqualTo, value: "5",
		},
		"shouldReturnJsonForOperatorIsStringEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsStringEqualTo.json",
			operator: filter.CompareOperatorIsStringEqualTo, value: "5",
		},
		"shouldReturnJsonForOperatorIsStringNotEqualToSingleValue": {
			file:     "testdata/And/singleValue/IsStringNotEqualTo.json",
			operator: filter.CompareOperatorIsStringNotEqualTo, value: "5",
		},
	}
	for name, tc := range mixedTests {
		t.Run(name, func(t *testing.T) {
			// given
			query := testBoolQueryBuilderWrapper{}
			query.BoolQueryBuilder = NewBoolQueryBuilder(&querySettings)

			// when
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

			// then
			require.NoError(t, err)
			json, err := query.toJson()
			require.NoError(t, err)

			expectedJson := testFolder.NewTestFolder().
				GetContent(t, tc.file)
			assert.JSONEq(t, expectedJson, json)
		})
	}
}

func TestFilterQueryOperatorOrMultiValue(t *testing.T) {
	mixedTests := map[string]struct {
		file     string
		operator filter.CompareOperator
		value    any
	}{
		// multi value
		"shouldReturnJsonForOperatorBeginsWithMultiValue": {
			file: "testdata/Or/multiValue/BeginsWith.json", operator: filter.CompareOperatorBeginsWith, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorContainsMultiValue": {
			file: "testdata/Or/multiValue/Contains.json", operator: filter.CompareOperatorContains,
			value: []interface{}{"test1", "test2"},
		},
		"shouldReturnJsonForOperatorDoesNotContainMultiValue": {
			file:     "testdata/Or/multiValue/DoesNotContain.json",
			operator: filter.CompareOperatorDoesNotContain,
			value:    []interface{}{"test1", "test2"},
		},
		"shouldReturnJsonForOperatorIsEqualToMultiValue": {
			file: "testdata/Or/multiValue/IsEqualTo.json", operator: filter.CompareOperatorIsEqualTo, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorIsNotEqualToMultiValue": {
			file:     "testdata/Or/multiValue/IsNotEqualTo.json",
			operator: filter.CompareOperatorIsNotEqualTo, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorIsStringEqualToMultiValue": {
			file:     "testdata/Or/multiValue/IsStringEqualTo.json",
			operator: filter.CompareOperatorIsStringEqualTo, value: []interface{}{5, 6},
		},
		"shouldReturnJsonForOperatorIsStringNotEqualToMultiValue": {
			file:     "testdata/Or/multiValue/IsStringNotEqualTo.json",
			operator: filter.CompareOperatorIsStringNotEqualTo, value: []interface{}{5, 6},
		},
		// single value
		"shouldReturnJsonForOperatorAfterDate": {
			file:     "testdata/Or/singleValue/AfterDate.json",
			operator: filter.CompareOperatorAfterDate, value: "2024-08-23T00:00:00.000+02:00",
		},
		"shouldReturnJsonForOperatorBeforeDate": {
			file:     "testdata/Or/singleValue/BeforeDate.json",
			operator: filter.CompareOperatorBeforeDate, value: "2024-08-23T00:00:00.000+02:00",
		},
		"shouldReturnJsonForOperatorBeginsWithSingleValue": {
			file: "testdata/Or/singleValue/BeginsWith.json", operator: filter.CompareOperatorBeginsWith, value: 5,
		},
		"shouldReturnJsonForOperatorContainsSingleValue": {
			file: "testdata/Or/singleValue/Contains.json", operator: filter.CompareOperatorContains, value: "test",
		},
		"shouldReturnJsonForOperatorDoesNotContainSingleValue": {
			file:     "testdata/Or/singleValue/DoesNotContain.json",
			operator: filter.CompareOperatorDoesNotContain, value: "test1",
		},
		"shouldReturnJsonForOperatorIsEqualToSingleValue": {
			file: "testdata/Or/singleValue/IsEqualTo.json", operator: filter.CompareOperatorIsEqualTo, value: "test",
		},
		"shouldReturnJsonForOperatorIsGreaterThan": {
			file:     "testdata/Or/singleValue/IsGreaterThan.json",
			operator: filter.CompareOperatorIsGreaterThan, value: 5,
		},
		"shouldReturnJsonForOperatorIsGreaterThanOrEqualTo": {
			file:     "testdata/Or/singleValue/IsGreaterThanOrEqualTo.json",
			operator: filter.CompareOperatorIsGreaterThanOrEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsIpEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsIpEqualTo.json",
			operator: filter.CompareOperatorIsIpEqualTo, value: "127.0.0.1",
		},
		"shouldReturnJsonForOperatorIsIpNotEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsIpNotEqualTo.json",
			operator: filter.CompareOperatorIsIpNotEqualTo, value: "127.0.0.1",
		},
		"shouldReturnJsonForOperatorIsLessThan": {
			file:     "testdata/Or/singleValue/IsLessThan.json",
			operator: filter.CompareOperatorIsLessThan, value: 5,
		},
		"shouldReturnJsonForOperatorIsLessThanOrEqualTo": {
			file:     "testdata/Or/singleValue/IsLessThanOrEqualTo.json",
			operator: filter.CompareOperatorIsLessThanOrEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsNotEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsNotEqualTo.json",
			operator: filter.CompareOperatorIsNotEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsNumberEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsNumberEqualTo.json",
			operator: filter.CompareOperatorIsNumberEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsNumberNotEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsNumberNotEqualTo.json",
			operator: filter.CompareOperatorIsNumberNotEqualTo, value: 5,
		},
		"shouldReturnJsonForOperatorIsStringEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsStringEqualTo.json",
			operator: filter.CompareOperatorIsStringEqualTo, value: "test",
		},
		"shouldReturnJsonForOperatorIsStringNotEqualToSingleValue": {
			file:     "testdata/Or/singleValue/IsStringNotEqualTo.json",
			operator: filter.CompareOperatorIsStringNotEqualTo, value: "5",
		},
	}
	for name, tc := range mixedTests {
		t.Run(name, func(t *testing.T) {
			// given
			query := testBoolQueryBuilderWrapper{}
			query.BoolQueryBuilder = NewBoolQueryBuilder(&querySettings)

			// when
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

			// then
			require.NoError(t, err)
			json, err := query.toJson()
			require.NoError(t, err)

			assert.JSONEq(t, testFolder.NewTestFolder().
				GetContent(t, tc.file), json)
		})
	}
}
