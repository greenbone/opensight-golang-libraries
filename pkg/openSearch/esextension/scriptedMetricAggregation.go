// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

import "github.com/aquasecurity/esquery"

// ScriptedMetricAggregation represents a scripted_metric aggregation for Elasticsearch.
type ScriptedMetricAggregation struct {
	name          string
	initScript    string
	mapScript     string
	combineScript string
	reduceScript  string
}

// Name returns the name of the ScriptedMetricAggregation, needed for the esquery.Aggregation interface.
func (a *ScriptedMetricAggregation) Name() string {
	return a.name
}

// Map returns a map representation of the ScriptedMetricAggregation, thus implementing the esquery.Mappable interface.
// Used for serialization to JSON.
func (a *ScriptedMetricAggregation) Map() map[string]interface{} {
	return map[string]interface{}{
		"scripted_metric": map[string]interface{}{
			"init_script":    a.initScript,
			"map_script":     a.mapScript,
			"combine_script": a.combineScript,
			"reduce_script":  a.reduceScript,
		},
	}
}

// ScriptedMetricAgg is a function that creates a new instance of ScriptedMetricAggregation.
// It takes the name, init script, map script, combine script, and reduce script as parameters and
// returns a pointer to the ScriptedMetricAggregation struct.
//
// Example usage:
//
//	a := ScriptedMetricAgg("unique_asset_ids", initScript, mapScript, combineScript, reduceScript)
func ScriptedMetricAgg(
	name string,
	initScript string,
	mapScript string,
	combineScript string,
	reduceScript string,
) *ScriptedMetricAggregation {
	return &ScriptedMetricAggregation{
		name:          name,
		initScript:    initScript,
		mapScript:     mapScript,
		combineScript: combineScript,
		reduceScript:  reduceScript,
	}
}

var _ esquery.Aggregation = &ScriptedMetricAggregation{}
