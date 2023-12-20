// SPDX-FileCopyrightText: 2023 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import "github.com/greenbone/opensight-golang-libraries/pkg/query/filter"

// FilterOption hold the information for a filter option. It can be used by a client to determine possible filters.
//
// Name: The name of the option
// Control: The type of control for the option
// Operators: The list of comparison operators for the option
// Values: The possible values for the option
// MultiSelect: Indicates whether the option supports multiple selections
type FilterOption struct {
	Name        filter.ReadableValue[string]                   `json:"name" binding:"required"`
	Control     filter.RequestOptionType                       `json:"control" binding:"required"`
	Operators   []filter.ReadableValue[filter.CompareOperator] `json:"operators" binding:"required"`
	Values      []string                                       `json:"values,omitempty"`
	MultiSelect bool                                           `json:"multiSelect" binding:"required"`
}
