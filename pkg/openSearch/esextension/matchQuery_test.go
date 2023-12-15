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
