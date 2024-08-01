// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package paging

// Response represents a response object containing information about pagination and total count of records.
//   - PageIndex: The index of the page (starting from 0).
//   - PageSize: The number of records per page.
type Response struct {
	PageIndex     int    `json:"index"`
	PageSize      int    `json:"size"`
	TotalRowCount uint64 `json:"total"`
}

func NewResponse(request *Request, totalRowCount uint64) *Response {
	if request == nil {
		return nil
	}

	return &Response{
		PageIndex:     request.PageIndex,
		PageSize:      request.PageSize,
		TotalRowCount: totalRowCount,
	}
}
