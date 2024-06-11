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
	// field mapping from filter field to database col name
	fieldMapping := map[string]string{
		"status":             "status_col_name",
		"source_id":          "source_id_col_name",
		"other_filter_field": "other_filter_field_col_name",
		"started":            "started_col_name",
		"severity":           "severity_col_name",
		"published":          "published_col_name",
	}

	tests := []struct {
		name      string
		mockArg   query.ResultSelector
		wantQuery string
		wantArgs  []any
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
			wantQuery: `WHERE "status_col_name" IN (?, ?) OR "source_id_col_name" IN (?, ?, ?) ORDER BY ? DESC OFFSET 2 LIMIT 5`,
			wantArgs:  []any{"invalid status", "valid status", "some_source_id", "another_source_id", "third_source_id", "started_col_name"},
		},
		{
			name: "build query with filter only",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "status",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    "invalid status",
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
			wantQuery: `WHERE "status_col_name" = ? AND "source_id_col_name" IN (?)`,
			wantArgs:  []any{"invalid status", "some_source_id"},
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
			wantQuery: `WHERE "status_col_name" IN (?) AND "source_id_col_name" IN (?) OFFSET 2 LIMIT 5`,
			wantArgs:  []any{"invalid status", "some_source_id"},
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
			wantQuery: `WHERE "status_col_name" IN (?) OFFSET 2 LIMIT 5`,
			wantArgs:  []any{"invalid status"},
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
			wantQuery: `WHERE "status_col_name" IN (?, ?) OFFSET 2 LIMIT 5`,
			wantArgs:  []any{"invalid status", "another status"},
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
			wantQuery: `WHERE "status_col_name" IN (?, ?) OR "source_id_col_name" IN (?, ?, ?) OR "other_filter_field_col_name" IN (?, ?, ?) ORDER BY ? DESC OFFSET 2 LIMIT 5`,
			wantArgs: []any{
				"invalid status", "valid status", "some_source_id", "another_source_id",
				"third_source_id", "some_field", "another_field", "third_field", "started_col_name",
			},
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
			wantQuery: `WHERE "status_col_name" IN (?, ?) OR "source_id_col_name" IN (?, ?, ?) OR "other_filter_field_col_name" IN (?, ?, ?) ORDER BY ? DESC OFFSET 2 LIMIT 5`,
			wantArgs: []any{
				"invalid status", "valid status", "some_source_id", "another_source_id",
				"third_source_id", "some_field", "another_field", "third_field", "started_col_name",
			},
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
			wantQuery: `WHERE "status_col_name" NOT IN (?, ?) OFFSET 2 LIMIT 5`,
			wantArgs:  []any{"invalid status", "another status"},
		},
		{
			name: "build valid query without filter fields",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields:   []filter.RequestField{},
					Operator: filter.LogicOperatorAnd,
				},
				Paging: &paging.Request{
					PageIndex: 3,
					PageSize:  10,
				},
				Sorting: &sorting.Request{
					SortColumn:    "started",
					SortDirection: "asc",
				},
			},
			wantQuery: " ORDER BY ? ASC OFFSET 3 LIMIT 10",
			wantArgs:  []any{"started_col_name"},
		},
		{
			name: "build valid query without filter object",
			mockArg: query.ResultSelector{
				Paging: &paging.Request{
					PageIndex: 3,
					PageSize:  10,
				},
				Sorting: &sorting.Request{
					SortColumn:    "started",
					SortDirection: "asc",
				},
			},
			wantQuery: " ORDER BY ? ASC OFFSET 3 LIMIT 10",
			wantArgs:  []any{"started_col_name"},
		},
		{
			name: "build query with filter paging, sorting, and date compare operators",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "started",
							Operator: filter.CompareOperatorAfterDate,
							Value:    "another_date",
						},
						{
							Name:     "started",
							Operator: filter.CompareOperatorBeforeDate,
							Value:    "some_date",
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
			wantQuery: `WHERE date_trunc('day'::text, "started_col_name") > date_trunc('day'::text, ?::timestamp) OR date_trunc('day'::text, "started_col_name") < date_trunc('day'::text, ?::timestamp) ORDER BY ? DESC OFFSET 2 LIMIT 5`,
			wantArgs:  []any{"another_date", "some_date", "started_col_name"},
		},
		{
			name: "build query with filter paging, sorting, greater than && less than operators",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "severity",
							Operator: filter.CompareOperatorIsGreaterThan,
							Value:    5.3,
						},
						{
							Name:     "severity",
							Operator: filter.CompareOperatorIsLessThan,
							Value:    8.2,
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
			wantQuery: `WHERE "severity_col_name" > ? OR "severity_col_name" < ? ORDER BY ? DESC OFFSET 2 LIMIT 5`,
			wantArgs:  []any{5.3, 8.2, "started_col_name"},
		},
		{
			name: "build query with filter paging, sorting, greater than && less than operators",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "severity",
							Operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
							Value:    5.3,
						},
						{
							Name:     "severity",
							Operator: filter.CompareOperatorIsLessThanOrEqualTo,
							Value:    8.2,
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
			wantQuery: `WHERE "severity_col_name" >= ? OR "severity_col_name" <= ? ORDER BY ? DESC OFFSET 2 LIMIT 5`,
			wantArgs:  []any{5.3, 8.2, "started_col_name"},
		},
		{
			name: "build query with filter paging, sorting, begins with && contains compare operators",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "other_filter_field",
							Operator: filter.CompareOperatorBeginsWith,
							Value:    []any{"some text"},
						},
						{
							Name:     "other_filter_field",
							Operator: filter.CompareOperatorContains,
							Value:    []any{"another text"},
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
			wantQuery: `WHERE ("other_filter_field_col_name" ILIKE ? || '%') OR ("other_filter_field_col_name" ILIKE '%' || ? || '%') ORDER BY ? DESC OFFSET 2 LIMIT 5`,
			wantArgs:  []any{"some text", "another text", "started_col_name"},
		},
		{
			name: "build query with filter paging, sorting, and date operators",
			mockArg: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "published",
							Operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
							Value:    5.3,
						},
						{
							Name:     "severity",
							Operator: filter.CompareOperatorIsLessThanOrEqualTo,
							Value:    8.2,
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
			wantQuery: `WHERE "published_col_name" >= ? OR "severity_col_name" <= ? ORDER BY ? DESC OFFSET 2 LIMIT 5`,
			wantArgs:  []any{5.3, 8.2, "started_col_name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			querySettings := &Settings{
				FilterFieldMapping: fieldMapping,
			}
			queryBuilder := NewPostgresQueryBuilder(querySettings)
			queryString, arg, err := queryBuilder.Build(tt.mockArg)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.wantArgs, arg)
			assert.Equal(t, queryString, tt.wantQuery)
		})
	}
}
