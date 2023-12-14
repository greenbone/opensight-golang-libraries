// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

import (
	"encoding/json"
	"reflect"
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
						{"n1": map[string]interface{}{"terms": map[string]interface{}{"field": "f1"}}},
						{"n2": map[string]interface{}{"terms": map[string]interface{}{"field": "f2"}}},
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

// copied from github.com/aquasecurity/esquery@v0.2.0 es_test.go
type mapTest struct {
	name     string
	given    esquery.Mappable
	expected map[string]interface{}
}

// copied from github.com/aquasecurity/esquery@v0.2.0 es_test.go
func runMapTests(t *testing.T, tests []mapTest) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			m := test.given.Map()

			// then
			// convert both maps to JSON in order to compare them. we do not
			// use reflect.DeepEqual on the maps as this doesn't always work
			exp, got, ok := sameJSON(test.expected, m)
			if !ok {
				t.Errorf("expected %s, got %s", exp, got)
			}
		})
	}
}

// copied from github.com/aquasecurity/esquery@v0.2.0 es_test.go
func sameJSON(a, b map[string]interface{}) (aJSON, bJSON []byte, ok bool) {
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)

	if aErr != nil || bErr != nil {
		return aJSON, bJSON, false
	}

	ok = reflect.DeepEqual(aJSON, bJSON)
	return aJSON, bJSON, ok
}
