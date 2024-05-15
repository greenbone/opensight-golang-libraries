// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/query"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder(t *testing.T) {
	//field mapping from filter field to database col name
	fieldMapping := map[string]string{
		"status":             "status_col_name",
		"source_id":          "source_id_col_name",
		"other_filter_field": "other_filter_field_col_name",
	}

	tests := []struct {
		name      string
		mockArg   query.ResultSelector
		wantQuery string
	}{
		{
			name: "build query with filter paging and sorting",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"invalid status", "valid status"},
						},
						{
							Name:     "source_id",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"some_source_id", "another_source_id", "third_source_id"},
						},
					},
					Operator: filter.LogicOperatorOr,
				},
				Paging: &paging.Request{
					PageIndex: 2,
					PageSize:  5,
				},
				Sorting: &sorting.Request{
					SortColumn:    "started",
					SortDirection: "desc",
				},
			},
			wantQuery: "WHERE \"status_col_name\" IN ('invalid status', 'valid status') OR \"source_id_col_name\" IN ('some_source_id', 'another_source_id', 'third_source_id') ORDER BY started DESC OFFSET 2 LIMIT 5",
		},
		{
			name: "build query with filter only",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"invalid status"},
						},
						{
							Name:     "source_id",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"some_source_id"},
						},
					},
					Operator: filter.LogicOperatorAnd,
				},
			},
			wantQuery: "WHERE \"status_col_name\" IN ('invalid status') AND \"source_id_col_name\" IN ('some_source_id')",
		},
		{
			name: "build query with filter and paging",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"invalid status"},
						},
						{
							Name:     "source_id",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"some_source_id"},
						},
					},
					Operator: filter.LogicOperatorAnd,
				},
				Paging: &paging.Request{
					PageIndex: 2,
					PageSize:  5,
				},
			},
			wantQuery: "WHERE \"status_col_name\" IN ('invalid status') AND \"source_id_col_name\" IN ('some_source_id') OFFSET 2 LIMIT 5",
		},
		{
			name: "build query with just one filter and paging",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"invalid status"},
						},
					},
					Operator: filter.LogicOperatorOr,
				},
				Paging: &paging.Request{
					PageIndex: 2,
					PageSize:  5,
				},
			},
			wantQuery: "WHERE \"status_col_name\" IN ('invalid status') OFFSET 2 LIMIT 5",
		},
		{
			name: "build query with just one filter with multiple values and paging",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"invalid status", "another status"},
						},
					},
					Operator: filter.LogicOperatorOr,
				},
				Paging: &paging.Request{
					PageIndex: 2,
					PageSize:  5,
				},
			},
			wantQuery: "WHERE \"status_col_name\" IN ('invalid status', 'another status') OFFSET 2 LIMIT 5",
		},
		{
			name: "build query with more than two filter paging and sorting",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"invalid status", "valid status"},
						},
						{
							Name:     "source_id",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"some_source_id", "another_source_id", "third_source_id"},
						},
						{
							Name:     "other_filter_field",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"some_field", "another_field", "third_field"},
						},
					},
					Operator: filter.LogicOperatorOr,
				},
				Paging: &paging.Request{
					PageIndex: 2,
					PageSize:  5,
				},
				Sorting: &sorting.Request{
					SortColumn:    "started",
					SortDirection: "desc",
				},
			},
			wantQuery: "WHERE \"status_col_name\" IN ('invalid status', 'valid status') \"source_id_col_name\" IN ('some_source_id', 'another_source_id', 'third_source_id') OR \"other_filter_field_col_name\" IN ('some_field', 'another_field', 'third_field') ORDER BY started DESC OFFSET 2 LIMIT 5",
		},
		{
			name: "build query with more than two filter paging and sorting",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"invalid status", "valid status"},
						},
						{
							Name:     "source_id",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"some_source_id", "another_source_id", "third_source_id"},
						},
						{
							Name:     "other_filter_field",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{"some_field", "another_field", "third_field"},
						},
					},
					Operator: filter.LogicOperatorOr,
				},
				Paging: &paging.Request{
					PageIndex: 2,
					PageSize:  5,
				},
				Sorting: &sorting.Request{
					SortColumn:    "started",
					SortDirection: "desc",
				},
			},
			wantQuery: "WHERE \"status_col_name\" IN ('invalid status', 'valid status') \"source_id_col_name\" IN ('some_source_id', 'another_source_id', 'third_source_id') OR \"other_filter_field_col_name\" IN ('some_field', 'another_field', 'third_field') ORDER BY started DESC OFFSET 2 LIMIT 5",
		},
		{
			name: "build query with just one filter with multiple values, compareOperatorNotEqualTo, and paging",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsNotEqualTo,
							Value:    []any{"invalid status", "another status"},
						},
					},
					Operator: filter.LogicOperatorOr,
				},
				Paging: &paging.Request{
					PageIndex: 2,
					PageSize:  5,
				},
			},
			wantQuery: "WHERE \"status_col_name\" NOT IN ('invalid status', 'another status') OFFSET 2 LIMIT 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			querySettings := &Settings{
				FilterFieldMapping: fieldMapping,
			}
			queryBuilder := NewPostgresQueryBuilder(querySettings)
			queryString := queryBuilder.Build(tt.mockArg)

			assert.Equal(t, queryString, tt.wantQuery)
		})
	}
}
