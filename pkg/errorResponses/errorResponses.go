// SPDX-FileCopyrightText: 2023 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package errorResponses provides rest api models for errors
package errorResponses

import "fmt"

type ErrorResponse struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Details string `json:"details,omitempty"`
}

type FieldErrorResponse struct {
	Type    string            `json:"type"` // always `validation-error`
	Title   string            `json:"title"`
	Details string            `json:"details,omitempty"`
	Errors  map[string]string `json:"errors"`
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
		Title: fmt.Sprintln(message...),
	}
}
