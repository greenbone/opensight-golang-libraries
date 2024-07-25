// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package sorting

import "strings"

const (
	DirectionDescending SortDirection = "desc"
	DirectionAscending  SortDirection = "asc"
	NoDirection         SortDirection = ""
)

func (s SortDirection) String() string {
	if s == NoDirection {
		return ""
	}
	return strings.ToUpper(string(s))
}

func SortDirectionFromString(str string) SortDirection {
	switch strings.ToUpper(str) {
	case DirectionDescending.String():
		return DirectionDescending
	case DirectionAscending.String():
		return DirectionAscending
	default:
		return NoDirection
	}
}

type SortDirection string
