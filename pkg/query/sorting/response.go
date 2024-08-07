// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package sorting

// Response represents the response structure for sorting column and direction.
// SortingColumn stores the name of the column which was used for sorting.
// SortingDirection stores the direction which was applied by the sorting.
type Response struct {
	SortingColumn    string        `json:"column"`
	SortingDirection SortDirection `json:"direction"`
}
