// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter

type Request struct {
	Operator LogicOperator  `json:"operator" binding:"required"`
	Fields   []RequestField `json:"fields" binding:"dive"`
}

// RequestOption configures a field for validation
type RequestOption struct {
	Name        ReadableValue[string]
	Control     RequestOptionType
	Operators   []ReadableValue[CompareOperator]
	Values      []string
	MultiSelect bool
}

type ReadableValue[T any] struct {
	// Label is the human-readable form of the value
	Label string `json:"label"`
	// Value is the value for the backend
	Value T `json:"value"`
}

type RequestOptionType struct {
	Type ControlType `json:"type" enums:"string,float,integer,enum"`
}

type RequestField struct {
	Name     string          `json:"name" binding:"required"`
	Keys     []string        `json:"keys"`
	Operator CompareOperator `json:"operator" binding:"required"`
	// Value can be a list of values or a value
	Value any `json:"value" binding:"required"`
}
