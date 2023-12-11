package paging

// validatePagingRequest validates a paging request.
func validatePagingRequest(req *Request) error {
	if req == nil {
		return &Error{Msg: "request is nil"}
	}

	if req.PageIndex < 0 {
		return &Error{Msg: "page index must be 0 or greater than 0"}
	}

	if req.PageSize <= 0 {
		return &Error{Msg: "page size must be greater than 0"}
	}

	return nil
}
