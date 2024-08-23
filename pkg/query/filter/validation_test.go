// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestOptionValidation(t *testing.T) {
	var requestOptions []RequestOption

	setup := func(t *testing.T) {
		requestOptions = []RequestOption{
			{
				Name: ReadableValue[string]{
					Value: "optionNameOne",
				},
				Control: RequestOptionType{
					Type: ControlTypeString,
				},
				Operators: []ReadableValue[CompareOperator]{
					{
						Value: CompareOperatorContains,
					},
					{
						Value: CompareOperatorDoesNotContain,
					},
				},
				MultiSelect: true,
			},
			{
				Name: ReadableValue[string]{
					Value: "optionNameTwo",
				},
				Control: RequestOptionType{
					Type: ControlTypeInteger,
				},
				Operators: []ReadableValue[CompareOperator]{
					{
						Value: CompareOperatorIsStringEqualTo,
					},
					{
						Value: CompareOperatorIsStringNotEqualTo,
					},
				},
				MultiSelect: false,
			},
		}
	}

	t.Run("shouldAllowUnsetFilter", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(nil, requestOptions)
		require.NoError(t, err)
	})

	t.Run("shouldAllowEmptyFilter", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields:   []RequestField{},
		}, requestOptions)
		require.NoError(t, err)
	})

	t.Run("shouldReturnErrorOnInvalidValueDataType", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionNameOne",
					Operator: CompareOperatorContains,
					Value:    0,
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field 'optionNameOne' must be from type '[]string'")
	})

	t.Run("shouldReturnErrorOnInvalidCompareOperator", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionNameOne",
					Operator: "invalidOperator",
					Value:    []interface{}{"testValue"},
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field 'optionNameOne' can not have the operator 'invalidOperator'")
	})

	t.Run("shouldReturnErrorOnEmptyCompareOperator", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionNameOne",
					Operator: "",
					Value:    []interface{}{"testValue"},
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field 'optionNameOne' can not have the operator ''")
	})

	t.Run("shouldReturnErrorOnEmptyFieldName", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "",
					Operator: CompareOperatorContains,
					Value:    []interface{}{"testValue"},
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field name '' is invalid")
	})

	t.Run("shouldAllowEmptyStringAsValue", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionNameOne",
					Operator: "contains",
					Value:    []interface{}{""},
				},
			},
		}, requestOptions)
		require.NoError(t, err)
	})

	t.Run("shouldReturnErrorOnInvalidFieldName", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "invalidName",
					Operator: CompareOperatorContains,
					Value:    0,
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, "field name 'invalidName' is invalid", validationError.Error())
	})

	t.Run("shouldReturnNoErrorWithValidData", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionNameOne",
					Operator: CompareOperatorContains,
					Value:    []interface{}{"testValue"},
				},
				{
					Name:     "optionNameTwo",
					Operator: CompareOperatorIsStringEqualTo,
					Value:    5.0,
				},
			},
		}, requestOptions)
		require.NoError(t, err)
	})
}

func TestCascadeRequestOption(t *testing.T) {
	var requestOptions []RequestOption

	setup := func(t *testing.T) {
		requestOptions = []RequestOption{
			{
				Name: ReadableValue[string]{
					Value: "tag",
				},
				Control: RequestOptionType{
					Type: ControlTypeAutocomplete,
				},
				Operators: []ReadableValue[CompareOperator]{
					{
						Value: CompareOperatorIsStringEqualTo,
					},
					{
						Value: CompareOperatorExists,
					},
				},
				MultiSelect: false,
			},
		}
	}

	t.Run("shouldReturnNoErrorWithInvalidCascade", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "tag",
					Keys:     []string{"key1"},
					Operator: CompareOperatorIsStringEqualTo,
					Value:    "",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, "field 'tag' must not be empty", validationError.Error())
	})

	t.Run("shouldReturnNoErrorWithMissingName", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "tag",
					Keys:     []string{"key1"},
					Operator: CompareOperatorIsStringEqualTo,
					Value:    "",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, "field 'tag' must not be empty", validationError.Error())
	})

	t.Run("shouldReturnNoErrorWithEmptyString", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "tag",
					Operator: CompareOperatorIsStringEqualTo,
					Value:    "",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, "field 'tag' must not be empty", validationError.Error())
	})

	t.Run("shouldReturnErrorWithSpace", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "tag",
					Keys:     []string{"key1"},
					Operator: CompareOperatorIsStringEqualTo,
					Value:    " ",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, "field 'tag' must not be empty", validationError.Error())
	})

	t.Run("shouldReturnErrorWithKeysLongerThen1", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "tag",
					Keys:     []string{"key1", "key2", "key3", "key"},
					Operator: CompareOperatorExists,
					Value:    "yes",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, "field 'tag' number of keys must be 1", validationError.Error())
	})

	t.Run("shouldReturnErrorWithValueYesOrNo", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "tag",
					Keys:     []string{"key1"},
					Operator: CompareOperatorExists,
					Value:    "Location",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, "for the field 'tag' the value must be 'yes' or 'no'", validationError.Error())
	})

	t.Run("shouldReturnNoErrorWithExistsOnTag", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "tag",
					Keys:     []string{"key1"},
					Operator: CompareOperatorExists,
					Value:    "yes",
				},
			},
		}, requestOptions)
		assert.NoError(t, err)
	})

	t.Run("shouldReturnNoErrorWithValidData", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "tag",
					Operator: CompareOperatorIsStringEqualTo,
					Keys:     []string{"a"},
					Value:    "b",
				},
			},
		}, requestOptions)
		require.NoError(t, err)
	})
}

func TestEnumRequestOption(t *testing.T) {
	var requestOptions []RequestOption

	setup := func(t *testing.T) {
		requestOptions = []RequestOption{
			{
				Name: ReadableValue[string]{
					Value: "optionName",
				},
				Control: RequestOptionType{
					Type: ControlTypeEnum,
				},
				Operators: []ReadableValue[CompareOperator]{
					{
						Value: CompareOperatorContains,
					},
					{
						Value: CompareOperatorDoesNotContain,
					},
				},
				Values: []string{
					"First", "Second", "Third",
				},
				MultiSelect: true,
			},
		}
	}

	t.Run("shouldReturnErrorOnInvalidEnumValue", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionName",
					Operator: CompareOperatorContains,
					Value:    []interface{}{"First", "Second", "invalid"},
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field 'optionName' can not have the value 'invalid'")
	})
}

func TestStringRequestOption(t *testing.T) {
	var requestOptions []RequestOption

	setup := func(t *testing.T) {
		requestOptions = []RequestOption{
			{
				Name: ReadableValue[string]{
					Value: "optionName",
				},
				Control: RequestOptionType{
					Type: ControlTypeString,
				},
				Operators: []ReadableValue[CompareOperator]{
					{
						Value: CompareOperatorContains,
					},
					{
						Value: CompareOperatorDoesNotContain,
					},
				},
				MultiSelect: false,
			},
		}
	}

	t.Run("shouldReturnErrorOnInvalidValueDataType", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionName",
					Operator: CompareOperatorContains,
					Value:    0,
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field 'optionName' must be from type 'string'")
	})
}

func TestIntegerRequestOption(t *testing.T) {
	var requestOptions []RequestOption

	setup := func(t *testing.T) {
		requestOptions = []RequestOption{
			{
				Name: ReadableValue[string]{
					Value: "optionName",
				},
				Control: RequestOptionType{
					Type: ControlTypeInteger,
				},
				Operators: []ReadableValue[CompareOperator]{
					{
						Value: CompareOperatorContains,
					},
					{
						Value: CompareOperatorDoesNotContain,
					},
				},
				MultiSelect: false,
			},
		}
	}

	t.Run("shouldReturnErrorOnInvalidValueDataType", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionName",
					Operator: CompareOperatorContains,
					Value:    "0",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field 'optionName' must be from type 'integer'")
	})
}

func TestFloatRequestOption(t *testing.T) {
	var requestOptions []RequestOption

	setup := func(t *testing.T) {
		requestOptions = []RequestOption{
			{
				Name: ReadableValue[string]{
					Value: "optionName",
				},
				Control: RequestOptionType{
					Type: ControlTypeFloat,
				},
				Operators: []ReadableValue[CompareOperator]{
					{
						Value: CompareOperatorContains,
					},
					{
						Value: CompareOperatorDoesNotContain,
					},
				},
				MultiSelect: false,
			},
		}
	}

	t.Run("shouldReturnErrorOnInvalidValueDataType", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionName",
					Operator: CompareOperatorContains,
					Value:    "0",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field 'optionName' must be from type 'float'")
	})
}

func TestBoolRequestOption(t *testing.T) {
	var requestOptions []RequestOption

	setup := func(t *testing.T) {
		requestOptions = []RequestOption{
			{
				Name: ReadableValue[string]{
					Value: "optionName",
				},
				Control: RequestOptionType{
					Type: ControlTypeBool,
				},
				Operators: []ReadableValue[CompareOperator]{
					{
						Value: CompareOperatorContains,
					},
					{
						Value: CompareOperatorDoesNotContain,
					},
				},
				MultiSelect: false,
			},
		}
	}

	t.Run("shouldReturnErrorOnInvalidValueDataType", func(t *testing.T) {
		setup(t)

		err := ValidateFilter(&Request{
			Operator: LogicOperatorAnd,
			Fields: []RequestField{
				{
					Name:     "optionName",
					Operator: CompareOperatorContains,
					Value:    "0",
				},
			},
		}, requestOptions)
		var validationError *ValidationError
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, validationError.Error(), "field 'optionName' must be from type 'bool'")
	})
}
