// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter

import (
	"strings"
	"time"

	"errors"

	"github.com/google/uuid"
	"github.com/greenbone/opensight-golang-libraries/pkg/slices"
)

// ValidateFilter validates the filter in the request
func ValidateFilter(request *Request, requestOptions []RequestOption) error {
	if request == nil {
		return nil
	}

	for i, field := range request.Fields {
		fieldNameIsValid := false
		if field.Name == "tag" {
			err := validateTagValues(field)
			if err != nil {
				return err
			}
		}
		for _, requestOption := range requestOptions {
			// validate field name
			if field.Name == requestOption.Name.Value {
				fieldNameIsValid = true

				// validate field operator
				fieldCanHaveOperator := slices.ContainsLambda(requestOption.Operators, func(
					operator ReadableValue[CompareOperator],
				) bool {
					return field.Operator == operator.Value
				})
				if !fieldCanHaveOperator {
					return NewValidationErrorWithStack("field '%s' can not have the operator '%s'", field.Name, field.Operator)
				}

				// validate field value
				if requestOption.MultiSelect {
					if _, ok := field.Value.([]interface{}); !ok {
						return NewValidationErrorWithStack("field '%s' must be from type '[]%s'", field.Name, requestOption.Control.Type)
					}

					for _, fieldValueItem := range field.Value.([]interface{}) {
						err := validateFieldValueType(requestOption, field.Name, fieldValueItem)
						if err != nil {
							var uuidValidationError *UuidValidationError
							if errors.As(err, &uuidValidationError) {
								request.Fields[i].Value = "00000000-0000-0000-0000-000000000000"
								return err
							}

							return err
						}
					}
				} else {
					err := validateFieldValueType(requestOption, field.Name, field.Value)
					if err != nil {
						return err
					}
				}
			}
		}
		if !fieldNameIsValid {
			return NewValidationErrorWithStack("field name '%s' is invalid", field.Name)
		}
	}

	return nil
}

func validateTagValues(request RequestField) error {
	if len(strings.TrimSpace(request.Value.(string))) == 0 {
		return NewValidationErrorWithStack("field '%s' must not be empty", request.Name)
	}
	if request.Operator != CompareOperatorExists {
		if request.Keys == nil || len(request.Keys) != 1 {
			return NewValidationErrorWithStack("field '%s' number of keys must be 1", request.Name)
		}
		if request.Value == "" {
			return NewValidationErrorWithStack("field '%s' must have a value ", request.Name)
		}
	} else {
		// Check for CompareOperatorExists
		if request.Keys == nil || len(request.Keys) != 1 {
			return NewValidationErrorWithStack("field '%s' number of keys must be 1", request.Name)
		}

		if strings.ToLower(request.Value.(string)) != "yes" &&
			strings.ToLower(request.Value.(string)) != "no" {
			return NewValidationErrorWithStack("for the field '%s' the value must be 'yes' or 'no'", request.Name)
		}
	}
	return nil
}

func validateFieldValueType(requestOption RequestOption, fieldName string, fieldValue any) error {
	if requestOption.Control.Type == ControlTypeInteger {
		if _, ok := fieldValue.(float64); !ok {
			return NewValidationErrorWithStack("field '%s' must be from type '%s'", fieldName, requestOption.Control.Type)
		}
	} else if requestOption.Control.Type == ControlTypeFloat {
		if _, ok := fieldValue.(float64); !ok {
			return NewValidationErrorWithStack("field '%s' must be from type '%s'", fieldName, requestOption.Control.Type)
		}
	} else if requestOption.Control.Type == ControlTypeString ||
		requestOption.Control.Type == ControlTypeEnum || requestOption.Control.Type == ControlTypeUuid {
		if _, ok := fieldValue.(string); !ok {
			return NewValidationErrorWithStack("field '%s' must be from type '%s'", fieldName, requestOption.Control.Type)
		}

		if requestOption.Control.Type == ControlTypeEnum {
			fieldCanHaveValue := slices.Contains(requestOption.Values, fieldValue.(string))
			if !fieldCanHaveValue {
				return NewValidationErrorWithStack("field '%s' can not have the value '%s'", fieldName, fieldValue)
			}
		} else if requestOption.Control.Type == ControlTypeUuid {
			_, err := uuid.Parse(fieldValue.(string))
			if err != nil {
				return NewUuidValidationError("field '%s' has an invalid UUID: %v", fieldName, fieldValue)
			}
		}
	} else if requestOption.Control.Type == ControlTypeDateTime {
		if _, ok := fieldValue.(string); !ok {
			return NewValidationErrorWithStack("field '%s' must be from type '%s'", fieldName, requestOption.Control.Type)
		}
		if _, err := time.Parse(time.RFC3339, fieldValue.(string)); err != nil {
			return NewValidationErrorWithStack("field '%s' must contain a valid RFC3339 time", fieldName)
		}
	} else if requestOption.Control.Type == ControlTypeAutocomplete {
		if _, ok := fieldValue.(string); !ok {
			return NewValidationErrorWithStack("field '%s' must be from type '%s'", fieldName, requestOption.Control.Type)
		}
	} else {
		return NewValidationErrorWithStack("request option control type '%s' is not supported", requestOption.Control.Type)
	}
	return nil
}
