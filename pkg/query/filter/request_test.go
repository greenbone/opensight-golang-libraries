// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
)

func TestReadableValueCompare(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		run      func() int
	}{
		{
			name:     "enum",
			expected: 1,
			run: func() int {
				x := filter.NewReadableValue("aaa", filter.ControlTypeEnum)
				y := filter.NewReadableValue("bbb", filter.ControlTypeBool)
				return x.Cmp(y)
			},
		},
		{
			name:     "unorderable-value",
			expected: -1,
			run: func() int {
				x := filter.NewReadableValue("aaa", 1)
				y := filter.NewReadableValue("bbb", 0)
				return x.Cmp(y)
			},
		},
		{
			name:     "equal-values",
			expected: -1,
			run: func() int {
				x := filter.NewReadableValue("aaa", filter.ControlTypeEnum)
				y := filter.NewReadableValue("bbb", filter.ControlTypeEnum)
				return x.Cmp(y)
			},
		},
		{
			name:     "equal-values-and-labels",
			expected: 0,
			run: func() int {
				x := filter.NewReadableValue("aaa", filter.ControlTypeEnum)
				y := filter.NewReadableValue("aaa", filter.ControlTypeEnum)
				return x.Cmp(y)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.run())
		})
	}
}

func TestReadableValues(t *testing.T) {
	tests := []struct {
		name  string
		given []filter.ReadableValue[filter.ControlType]
		want  []filter.ReadableValue[filter.ControlType]
	}{
		{
			name:  "empty",
			given: []filter.ReadableValue[filter.ControlType]{},
			want:  []filter.ReadableValue[filter.ControlType]{},
		},
		{
			name:  "already-sorted",
			given: []filter.ReadableValue[filter.ControlType]{{Label: "aaa", Value: filter.ControlTypeBool}, {Label: "bbb", Value: filter.ControlTypeEnum}},
			want:  []filter.ReadableValue[filter.ControlType]{{Label: "aaa", Value: filter.ControlTypeBool}, {Label: "bbb", Value: filter.ControlTypeEnum}},
		},
		{
			name:  "reordered",
			given: []filter.ReadableValue[filter.ControlType]{{Label: "aaa", Value: filter.ControlTypeEnum}, {Label: "bbb", Value: filter.ControlTypeBool}},
			want:  []filter.ReadableValue[filter.ControlType]{{Label: "bbb", Value: filter.ControlTypeBool}, {Label: "aaa", Value: filter.ControlTypeEnum}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filter.SortedReadableValues(test.given...)
			require.Equal(t, test.want, got)
		})
	}
}
