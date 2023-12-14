package esextensions

import (
	"reflect"
	"testing"
)

func TestScriptedSumAggregation_Name(t *testing.T) {
	a := ScriptedSumAgg("testName", "testScript")
	if a.Name() != "testName" {
		t.Errorf("Expected name to be %s, but got %s", "testName", a.Name())
	}
}

func TestScriptedSumAggregation_Map(t *testing.T) {
	a := ScriptedSumAgg("testName", "testScript")
	expectedMap := map[string]interface{}{
		"sum": map[string]interface{}{
			"script": map[string]interface{}{
				"source": "testScript",
			},
		},
	}

	if !reflect.DeepEqual(a.Map(), expectedMap) {
		t.Errorf("Expected map to be %+v, but got %+v", expectedMap, a.Map())
	}
}

func TestScriptedSumAgg(t *testing.T) {
	a := ScriptedSumAgg("testName", "testScript")
	if a.name != "testName" || a.script != "testScript" {
		t.Errorf("Expected name and script to be %s and %s, but got %s and %s", "testName", "testScript", a.name, a.script)
	}
}
