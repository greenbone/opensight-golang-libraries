package paging

// Response represents a response object containing information about pagination and total count of records.
//   - PageIndex: The index of the page (starting from 0). This is required.
//   - PageSize: The number of records per page. This is required.
//   - TotalDisplayableResults: The total number of results that can be displayed. This is required.
//   - TotalResults: The total number of results, including those that may not be displayed. This is optional and must not be set if the value does not differ from `TotalDisplayableResults`
type Response struct {
	PageIndex               int    `json:"index" binding:"required"`
	PageSize                int    `json:"size" binding:"required"`
	TotalDisplayableResults uint64 `json:"totalDisplayableResults" binding:"required"`
	TotalResults            uint64 `json:"totalResults,omitempty"`
}

func NewResponse(request *Request, totalDisplayableResults uint64) *Response {
	if request == nil {
		return nil
	}

	return &Response{
		PageIndex:               request.PageIndex,
		PageSize:                request.PageSize,
		TotalDisplayableResults: totalDisplayableResults,
	}
}

func NewResponseWithTotalResults(request *Request, totalDisplayableResults, totalResults uint64) *Response {
	if request == nil {
		return nil
	}

	if totalDisplayableResults == totalResults {
		return NewResponse(request, totalDisplayableResults)
	}

	return &Response{
		PageIndex:               request.PageIndex,
		PageSize:                request.PageSize,
		TotalDisplayableResults: totalDisplayableResults,
		TotalResults:            totalResults,
	}
}
