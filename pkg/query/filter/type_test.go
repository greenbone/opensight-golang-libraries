// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
)

func TestEnumCompare(t *testing.T) {
	tests := []struct {
		name     string
		run      func() int
		expected int
	}{
		{
			name:     "ControlType/valid/valid/equal",
			run:      func() int { return filter.ControlTypeBool.Cmp(filter.ControlTypeBool) },
			expected: 0,
		},
		{
			name:     "ControlType/valid/valid",
			run:      func() int { return filter.ControlTypeBool.Cmp(filter.ControlTypeEnum) },
			expected: -1,
		},
		{
			name:     "ControlType/valid/invalid",
			run:      func() int { return filter.ControlTypeBool.Cmp("aaa") },
			expected: -1,
		},
		{
			name:     "ControlType/invalid/valid",
			run:      func() int { return filter.ControlType("aaa").Cmp(filter.ControlTypeEnum) },
			expected: 1,
		},
		{
			name:     "ControlType/invalid/invalid",
			run:      func() int { return filter.ControlType("bbb").Cmp("aaa") },
			expected: 1,
		},

		{
			name:     "LogicOperator/valid/valid/equal",
			run:      func() int { return filter.LogicOperatorOr.Cmp(filter.LogicOperatorOr) },
			expected: 0,
		},
		{
			name:     "LogicOperator/valid/valid",
			run:      func() int { return filter.LogicOperatorOr.Cmp(filter.LogicOperatorAnd) },
			expected: 1,
		},
		{
			name:     "LogicOperator/valid/invalid",
			run:      func() int { return filter.LogicOperatorOr.Cmp("aaa") },
			expected: -1,
		},
		{
			name:     "LogicOperator/invalid/valid",
			run:      func() int { return filter.LogicOperator("aaa").Cmp(filter.LogicOperatorAnd) },
			expected: 1,
		},
		{
			name:     "LogicOperator/invalid/invalid",
			run:      func() int { return filter.LogicOperator("bbb").Cmp("aaa") },
			expected: 1,
		},

		{
			name:     "CompareOperator/valid/valid/equal",
			run:      func() int { return filter.CompareOperatorBeginsWith.Cmp(filter.CompareOperatorBeginsWith) },
			expected: 0,
		},
		{
			name:     "CompareOperator/valid/valid",
			run:      func() int { return filter.CompareOperatorBeginsWith.Cmp(filter.CompareOperatorContains) },
			expected: -1,
		},
		{
			name:     "CompareOperator/valid/invalid",
			run:      func() int { return filter.CompareOperatorBeginsWith.Cmp("aaa") },
			expected: -1,
		},
		{
			name:     "CompareOperator/invalid/valid",
			run:      func() int { return filter.CompareOperator("aaa").Cmp(filter.CompareOperatorContains) },
			expected: 1,
		},
		{
			name:     "CompareOperator/invalid/invalid",
			run:      func() int { return filter.CompareOperator("bbb").Cmp("aaa") },
			expected: 1,
		},

		{
			name:     "AggregateMetric/valid/valid/equal",
			run:      func() int { return filter.AggregateMetricSum.Cmp(filter.AggregateMetricSum) },
			expected: 0,
		},
		{
			name:     "AggregateMetric/valid/valid",
			run:      func() int { return filter.AggregateMetricSum.Cmp(filter.AggregateMetricMin) },
			expected: -1,
		},
		{
			name:     "AggregateMetric/valid/invalid",
			run:      func() int { return filter.AggregateMetricSum.Cmp("aaa") },
			expected: -1,
		},
		{
			name:     "AggregateMetric/invalid/valid",
			run:      func() int { return filter.AggregateMetric("aaa").Cmp(filter.AggregateMetricMin) },
			expected: 1,
		},
		{
			name:     "AggregateMetric/invalid/invalid",
			run:      func() int { return filter.AggregateMetric("bbb").Cmp("aaa") },
			expected: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.run())
		})
	}
}
