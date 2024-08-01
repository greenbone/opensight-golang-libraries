// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package sorting

import "fmt"

// ValidateSortingRequest validates a sorting request.
func ValidateSortingRequest(req *Request) error {
	if req == nil {
		return &Error{Msg: "sorting request is nil"}
	}
	if req.SortColumn == "" {
		return &Error{Msg: "sorting column is empty"}
	}

	if req.SortDirection == "" {
		return &Error{Msg: "sorting direction is empty"}
	}

	if SortDirectionFromString(req.SortDirection.String()) == NoDirection {
		return &Error{
			Msg: fmt.Sprintf("%s is no valid sorting direction, possible values are asc, desc", req.SortDirection.String()),
		}
	}

	return nil
}
