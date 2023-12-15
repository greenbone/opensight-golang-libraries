package esextensions

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aquasecurity/esquery"
)

// mapTest is a helper struct for unit tests of objects that implement the esquery.Mappable interface.
// copied from github.com/aquasecurity/esquery@v0.2.0 es_test.go
type mapTest struct {
	name     string
	given    esquery.Mappable
	expected map[string]interface{}
}

// runMapTests is a helper function to run test cases following the mapTest structure.
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
