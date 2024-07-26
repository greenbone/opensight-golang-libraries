// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package paging

const (
	opensearchQueryLimit = 10000
)

// Response represents a response object containing information about pagination and total count of records.
//   - PageIndex: The index of the page (starting from 0). This is required.
//   - PageSize: The number of records per page. This is required.
//   - TotalDisplayableResults: The total number of results that can be paginated. Due to database restrictions, in case of large number of results, some of the results cannot be retrieved. In such cases, this number will be lower than the `TotalResults`. This is required.
//   - TotalResults: The total count of results as it exists in database, including those that may not be retrieved. This is optional and must not be set if the value does not differ from `TotalDisplayableResults`
type Response struct {
	PageIndex               int    `json:"index" binding:"required"`
	PageSize                int    `json:"size" binding:"required"`
	TotalDisplayableResults uint64 `json:"totalDisplayableResults" binding:"required"`
	TotalResults            uint64 `json:"totalResults,omitempty"`
}

func NewResponse(request *Request, totalResults uint64) *Response {
	if request == nil {
		return nil
	}

	if totalResults > opensearchQueryLimit {
		return &Response{
			PageIndex:               request.PageIndex,
			PageSize:                request.PageSize,
			TotalDisplayableResults: opensearchQueryLimit,
			TotalResults:            totalResults,
		}
	}

	return &Response{
		PageIndex:               request.PageIndex,
		PageSize:                request.PageSize,
		TotalDisplayableResults: totalResults,
	}
}
