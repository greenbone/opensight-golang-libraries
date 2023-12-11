package paging

// validatePagingRequest validates a paging request.
// It checks if the request is nil, if the page index is less than 0,
// and if the page size is less than or equal to 0.
// It returns an error if any of these conditions are met.
// Otherwise, it returns nil.
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
