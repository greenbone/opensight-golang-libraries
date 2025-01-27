// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package paging

import (
	"fmt"
)

type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return e.Msg
}

func NewPagingError(format string, value ...any) error {
	return &Error{
		Msg: fmt.Sprintf(format, value...),
	}
}
