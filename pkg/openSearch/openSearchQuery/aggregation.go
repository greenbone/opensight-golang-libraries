// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"github.com/aquasecurity/esquery"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
)

// Aggregation is an interface for all aggregations that can be converted to an esquery.Aggregation.
type Aggregation interface {
	// ToEsAggregation to es aggregation
	ToEsAggregation() esquery.Aggregation
}

// TermsAgg creates a new TermsAggregation.
//
// The name is the name of the aggregation.
// The field is the field to aggregate on.
func TermsAgg(name, field string) *TermsAggregation {
	return &TermsAggregation{
		name:  name,
		field: field,
		size:  0,
	}
}

// TermsAggregation represents an OpenSearch terms aggregation.
type TermsAggregation struct {
	name         string
	field        string
	size         uint64
	aggregations []Aggregation
}

// Size sets the size of the aggregation.
func (t *TermsAggregation) Size(size uint64) *TermsAggregation {
	t.size = size
	return t
}

// Aggs sets the sub-aggregations of the aggregation.
func (t *TermsAggregation) Aggs(aggregations ...Aggregation) *TermsAggregation {
	t.aggregations = aggregations
	return t
}

// ToEsAggregation converts the aggregation to an esquery.Aggregation.
func (t *TermsAggregation) ToEsAggregation() esquery.Aggregation {
	var esAggregations []esquery.Aggregation

	for _, aggregation := range t.aggregations {
		esAggregations = append(esAggregations, aggregation.ToEsAggregation())
	}

	aggregation := esquery.TermsAgg(t.name, t.field)

	if t.size > 0 {
		aggregation = aggregation.Size(t.size)
	}

	return aggregation.Aggs(esAggregations...)
}

// MetricAggregation represents an OpenSearch metric aggregation.
type MetricAggregation struct {
	name   string
	field  string
	metric filter.AggregateMetric
}

func metricAgg(name string, metric filter.AggregateMetric, field string) *MetricAggregation {
	return &MetricAggregation{
		name:   name,
		field:  field,
		metric: metric,
	}
}

// SumAgg creates a new OpenSearch sum aggregation.
func SumAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricSum, field)
}

// MinAgg creates a new OpenSearch min aggregation.
func MinAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricMin, field)
}

// MaxAgg creates a new OpenSearch max aggregation.
func MaxAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricMax, field)
}

// AvgAgg creates a new OpenSearch avg aggregation.
func AvgAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricAvg, field)
}

// ValueCountAgg creates a new OpenSearch valueCount aggregation.
func ValueCountAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricValueCount, field)
}

// ToEsAggregation converts the aggregation to an esquery.Aggregation.
func (m *MetricAggregation) ToEsAggregation() esquery.Aggregation {
	var aggregation esquery.Aggregation

	switch m.metric {
	case filter.AggregateMetricSum:
		aggregation = esquery.Sum(m.name, m.field)
	case filter.AggregateMetricMin:
		aggregation = esquery.Min(m.name, m.field)
	case filter.AggregateMetricMax:
		aggregation = esquery.Max(m.name, m.field)
	case filter.AggregateMetricAvg:
		aggregation = esquery.Avg(m.name, m.field)
	case filter.AggregateMetricValueCount:
		aggregation = esquery.ValueCount(m.name, m.field)
	}

	return aggregation
}
