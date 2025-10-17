// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package osquery

import (
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
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
