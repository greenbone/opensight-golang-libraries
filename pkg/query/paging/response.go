package paging

// Response represents a response object containing information about pagination and total count of records.
//   - PageIndex: The index of the page (starting from 0).
//   - PageSize: The number of records per page.
type Response struct {
	PageIndex               int    `json:"index" binding:"required"`
	PageSize                int    `json:"size" binding:"required"`
	TotalDisplayableResults uint64 `json:"totalDisplayableResults" binding:"required"`
	TotalResults            uint64 `json:"totalResults,omitempty"`
}

func NewResponse(request *Request, totalDisplayableResults, totalResults uint64) *Response {
	if request == nil {
		return nil
	}

	return &Response{
		PageIndex:               request.PageIndex,
		PageSize:                request.PageSize,
		TotalDisplayableResults: totalDisplayableResults,
		TotalResults:            totalResults,
	}
}
