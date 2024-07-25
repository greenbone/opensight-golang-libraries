// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package sorting

import (
	"fmt"
)

type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return e.Msg
}

func NewSortingError(format string, value ...any) error {
	return &Error{
		Msg: fmt.Sprintf(format, value...),
	}
}
