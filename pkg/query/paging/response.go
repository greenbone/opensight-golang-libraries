package paging

// Response represents a response object containing information about pagination and total count of records.
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
