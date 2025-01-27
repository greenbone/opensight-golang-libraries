// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

import (
	"testing"

	"github.com/aquasecurity/esquery"
)

func TestComposite(t *testing.T) {
	runMapTests(t, []mapTest{
		{
			name:  "only name",
			given: Composite("myComposite"),
			expected: map[string]interface{}{
				"composite": map[string]interface{}{},
			},
		},
		{
			name:  "size",
			given: Composite("myComposite").Size(2),
			expected: map[string]interface{}{
				"composite": map[string]interface{}{
					"size": 2,
				},
			},
		},
		{
			name:  "sources",
			given: Composite("myComposite").Sources(Terms("n1", "f1"), Terms("n2", "f2")),
			expected: map[string]interface{}{
				"composite": map[string]interface{}{
					"sources": []map[string]interface{}{
						{"n1": map[string]interface{}{"terms": map[string]interface{}{"field": "f1", "order": "asc"}}},
						{"n2": map[string]interface{}{"terms": map[string]interface{}{"field": "f2", "order": "asc"}}},
					},
				},
			},
		},
		{
			name:  "after",
			given: Composite("myComposite").After(map[string]string{"key": "value"}),
			expected: map[string]interface{}{
				"composite": map[string]interface{}{
					"after": map[string]interface{}{"key": "value"},
				},
			},
		},
		{
			name: "aggregations",
			given: Composite("myComposite").Aggregations(
				esquery.Avg("n1", "f1"),
				esquery.Cardinality("n2", "f2"),
			),
			expected: map[string]interface{}{
				"composite": map[string]interface{}{},
				"aggregations": map[string]interface{}{
					"n1": map[string]interface{}{
						"avg": map[string]interface{}{
							"field": "f1",
						},
					},
					"n2": map[string]interface{}{
						"cardinality": map[string]interface{}{
							"field": "f2",
						},
					},
				},
			},
		},
		{
			name:  "embedded in Search().Aggs(...)",
			given: esquery.Search().Aggs(Composite("myComposite")),
			expected: map[string]interface{}{
				"aggs": map[string]interface{}{
					"myComposite": map[string]interface{}{
						"composite": map[string]interface{}{},
					},
				},
			},
		},
	})
}
