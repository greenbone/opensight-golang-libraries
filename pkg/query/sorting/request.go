// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package sorting

// Request represents a sorting request with a specified sort column and sort direction.
//
// Fields:
// - SortColumn: the column to sort on
// - SortDirection: the direction of sorting (asc or desc)
type Request struct {
	SortColumn    string        `json:"column"`
	SortDirection SortDirection `json:"direction"`
}
