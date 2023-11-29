package esextensions

// ... other imports
import "github.com/aquasecurity/esquery"

type ScriptedSumAggregation struct {
	name   string
	script string
}

func (a *ScriptedSumAggregation) Name() string {
	return a.name
}

func (a *ScriptedSumAggregation) Map() map[string]interface{} {
	return map[string]interface{}{
		"sum": map[string]interface{}{
			"script": map[string]interface{}{
				"source": a.script,
			},
		},
	}
}

func ScriptedSumAgg(name string, script string) *ScriptedSumAggregation {
	return &ScriptedSumAggregation{
		name:   name,
		script: script,
	}
}

var _ esquery.Aggregation = &ScriptedSumAggregation{}
