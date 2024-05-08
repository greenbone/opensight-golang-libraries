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
	querySettings := &Settings{
		FilterFieldMapping: map[string]string{
			"status":    "status",
			"source_id": "source_id",
		},
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
			wantQuery: "WHERE status = 'invalid status' status = 'valid status' source_id = 'some_source_id' source_id = 'another_source_id' OR source_id = 'third_source_id' ORDER BY started DESC OFFSET 2 LIMIT 5",
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
			wantQuery: "WHERE status = 'invalid status' AND source_id = 'some_source_id' ",
		},
		{
			name: "Build query with filter and paging",
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
			wantQuery: "WHERE status = 'invalid status' AND source_id = 'some_source_id' OFFSET 2 LIMIT 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryBuilder := NewPostgresQueryBuilder(querySettings)
			queryString := queryBuilder.Build(tt.mockArg)

			assert.Equal(t, queryString, tt.wantQuery)
		})
	}
}
