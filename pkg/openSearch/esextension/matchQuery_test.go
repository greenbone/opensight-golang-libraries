// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

import (
	"testing"
)

func TestMatch(t *testing.T) {
	runMapTests(t, []mapTest{
		{
			name:  "match query map",
			given: Match("field_name", "value"),
			expected: map[string]interface{}{
				"match": map[string]interface{}{
					"field_name": "value",
				},
			},
		},
	})
}
