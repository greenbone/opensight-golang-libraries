package esextensions

// ... other imports
import "github.com/aquasecurity/esquery"

// ScriptedSumAggregation represents an aggregation that calculates the sum using a scripted expression.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.17/search-aggregations-metrics-sum-aggregation.html#_script_14 .
// To be used in conjunction with the esquery library https://github.com/aquasecurity/esquery
type ScriptedSumAggregation struct {
	name   string
	script string
}

// Name returns the name of the ScriptedSumAggregation, needed for the esquery.Aggregation interface.
func (a *ScriptedSumAggregation) Name() string {
	return a.name
}

// Map returns a map representation of the ScriptedSumAggregation, thus implementing the esquery.Mappable interface.
// Used for serialization to JSON.
func (a *ScriptedSumAggregation) Map() map[string]interface{} {
	return map[string]interface{}{
		"sum": map[string]interface{}{
			"script": map[string]interface{}{
				"source": a.script,
			},
		},
	}
}

// ScriptedSumAgg is a function that creates a new instance of ScriptedSumAggregation.
// It takes the name and script as parameters and returns a pointer to the ScriptedSumAggregation struct.
// Example usage:
//
//	a := ScriptedSumAgg("testName", "testScript")
func ScriptedSumAgg(name string, script string) *ScriptedSumAggregation {
	return &ScriptedSumAggregation{
		name:   name,
		script: script,
	}
}

var _ esquery.Aggregation = &ScriptedSumAggregation{}
