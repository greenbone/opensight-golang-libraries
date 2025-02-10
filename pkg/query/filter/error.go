// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter

import (
	"fmt"
)

type ValidationError struct {
	message string
}

func (v *ValidationError) Error() string {
	return v.message
}

func NewValidationError(format string, value ...any) *ValidationError {
	return &ValidationError{
		message: fmt.Sprintf(format, value...),
	}
}

type InvalidFilterFieldError struct {
	message string
}

func (i *InvalidFilterFieldError) Error() string {
	return i.message
}

func NewInvalidFilterFieldError(format string, value ...any) error {
	return &InvalidFilterFieldError{
		message: fmt.Sprintf(format, value...),
	}
}

type UuidValidationError struct {
	message string
}

func (v *UuidValidationError) Error() string {
	return v.message
}

func NewUuidValidationError(format string, value ...any) *UuidValidationError {
	return &UuidValidationError{
		message: fmt.Sprintf(format, value...),
	}
}
