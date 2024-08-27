// SPDX-FileCopyrightText: 2023-2024 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package errorResponses provides rest api models for errors
package errorResponses

import "fmt"

type ErrorResponse struct {
	Type    string            `json:"type"`
	Title   string            `json:"title"`
	Details string            `json:"details,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

const errorTypePrefix = "greenbone/"

const (
	ErrorTypeGeneric     = errorTypePrefix + "generic-error"
	ErrorTypeValidation  = errorTypePrefix + "validation-error"
	ErrorTypeUpstreamApi = errorTypePrefix + "upstream-api-error"
	ErrorTypeApi         = errorTypePrefix + "api-error"
	ErrorTypeConnection  = errorTypePrefix + "connection-error"
)

var ErrorInternalResponse = ErrorResponse{
	Type:  ErrorTypeGeneric,
	Title: "internal error",
}

// NewErrorGenericResponse returns a [ErrorResponse] of type Generic with the given error message. The message is handled the same as [fmt.Println].
func NewErrorGenericResponse(message ...any) ErrorResponse {
	return ErrorResponse{
		Type:  ErrorTypeGeneric,
		Title: fmt.Sprint(message...),
	}
}

// NewErrorValidationResponse returns a [ErrorResponse] of type Validation. `errors` is a mapping from field name to error message for this specific field.
func NewErrorValidationResponse(title, details string, errors map[string]string) ErrorResponse {
	return ErrorResponse{
		Type:    ErrorTypeValidation,
		Title:   title,
		Details: details,
		Errors:  errors,
	}
}
