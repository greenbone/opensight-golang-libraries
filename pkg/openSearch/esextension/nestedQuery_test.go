// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

import (
	"testing"

	"github.com/aquasecurity/esquery"
)

func TestNested(t *testing.T) {
	runMapTests(t, []mapTest{
		{
			name:  "nested query map",
			given: Nested("asset.tags", *esquery.Bool()),
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"path":  "asset.tags",
					"query": map[string]interface{}{"bool": map[string]interface{}{}},
				},
			},
		},
		{
			name:  "nested query map with additional path element",
			given: Nested("asset.tags.keyword", *esquery.Bool()),
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"path":  "asset.tags",
					"query": map[string]interface{}{"bool": map[string]interface{}{}},
				},
			},
		},
	})
}
