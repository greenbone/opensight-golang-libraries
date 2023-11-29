// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_query

import (
	"github.com/aquasecurity/esquery"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
)

type Aggregation interface {
	// ToEsAggregation to es aggregation
	ToEsAggregation() esquery.Aggregation
}

func TermsAgg(name, field string) *TermsAggregation {
	return &TermsAggregation{
		name:  name,
		field: field,
		size:  0,
	}
}

type TermsAggregation struct {
	name         string
	field        string
	size         uint64
	aggregations []Aggregation
}

func (t *TermsAggregation) Size(size uint64) *TermsAggregation {
	t.size = size
	return t
}

func (t *TermsAggregation) Aggs(aggregations ...Aggregation) *TermsAggregation {
	t.aggregations = aggregations
	return t
}

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

func SumAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricSum, field)
}

func MinAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricMin, field)
}

func MaxAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricMax, field)
}

func AvgAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricAvg, field)
}

func ValueCountAgg(name string, field string) *MetricAggregation {
	return metricAgg(name, filter.AggregateMetricValueCount, field)
}

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
