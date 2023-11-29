// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package escomposite

import "github.com/aquasecurity/esquery"

// CompositeAgg represents a composite aggregation, as described in
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/search-aggregations-bucket-composite-aggregation.html
// to be used in conjunction with the esquery library https://github.com/aquasecurity/esquery
type CompositeAgg struct {
	name         string
	size         uint64
	sources      []esquery.Mappable
	after        map[string]string
	aggregations []esquery.Aggregation
}

// Composite creates an aggregation of type "composite".
func Composite(name string) *CompositeAgg {
	return &CompositeAgg{
		name: name,
	}
}

// Name returns the name of the aggregation.
func (agg *CompositeAgg) Name() string {
	return agg.name
}

// Size sets the maximum number of buckets to return.
func (agg *CompositeAgg) Size(size uint64) *CompositeAgg {
	agg.size = size
	return agg
}

// Sources sets the sources for the buckets.
func (agg *CompositeAgg) Sources(sources ...esquery.Mappable) *CompositeAgg {
	agg.sources = append(agg.sources, sources...)
	return agg
}

// After sets the identification for the entry after which the next results should be returned.
func (agg *CompositeAgg) After(after map[string]string) *CompositeAgg {
	agg.after = after
	return agg
}

// Aggregations sets the aggregations to be used for the buckets.
func (agg *CompositeAgg) Aggregations(aggregations ...esquery.Aggregation) *CompositeAgg {
	agg.aggregations = append(agg.aggregations, aggregations...)
	return agg
}

// Map returns a map representation of the aggregation, thus implementing the
// Mappable interface.
func (agg *CompositeAgg) Map() map[string]interface{} {
	compositeMap := make(map[string]interface{})

	if agg.size > 0 {
		compositeMap["size"] = agg.size
	}

	if len(agg.sources) > 0 {
		sources := make([]map[string]interface{}, len(agg.sources))
		for i := range agg.sources {
			sources[i] = agg.sources[i].Map()
		}
		compositeMap["sources"] = sources
	}

	if agg.after != nil {
		compositeMap["after"] = agg.after
	}

	result := map[string]interface{}{
		"composite": compositeMap,
	}

	if len(agg.aggregations) > 0 {
		aggregationMap := make(map[string]interface{})

		for _, a := range agg.aggregations {
			aggregationMap[a.Name()] = a.Map()
		}

		result["aggregations"] = aggregationMap
	}

	return result
}
