// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter

// Request is a struct representing a filter request.
// Operator is the logic operator used for the request.
// Fields is a slice of RequestField, representing the fields to be used for the filtering.
type Request struct {
	Operator LogicOperator  `json:"operator" binding:"required"`
	Fields   []RequestField `json:"fields" binding:"dive"`
}

// RequestOption configures a field for validation
//
// Name: The name of the option
// Control: The type of control for the option
// Operators: The list of comparison operators for the option
// Values: The possible values for the option
// MultiSelect: Indicates whether the option supports multiple selections
type RequestOption struct {
	Name        ReadableValue[string]
	Control     RequestOptionType
	Operators   []ReadableValue[CompareOperator]
	Values      []string
	MultiSelect bool
}

// ReadableValue is a generic type that represents a human-readable value with a corresponding backend value.
// It has two fields: `Label` (the human-readable form of the value) and `Value` (the value for the backend).
type ReadableValue[T any] struct {
	// Label is the human-readable form of the value
	Label string `json:"label"`
	// Value is the value for the backend
	Value T `json:"value"`
}

// RequestOptionType configures the type of control for a field in a request option.
type RequestOptionType struct {
	Type ControlType `json:"type" enums:"string,float,integer,enum,bool"`
}

// RequestField represents a field in a request
// Field Name: The name of the field
// Field Keys: Sequence of keys of a nested key structure - only used for fields with a nested structure. Example: Tag -> Name: ABC (which would be represented as []string{"Tag", "Name: ABC"} )
// Field Operator: The comparison operator for the field
// Field Value: The value of the field, which can be a list of values or a single value
type RequestField struct {
	Name     string          `json:"name" binding:"required"`
	Keys     []string        `json:"keys"`
	Operator CompareOperator `json:"operator" binding:"required"`
	// Value can be a list of values or a value
	Value any `json:"value" binding:"required"`
}
