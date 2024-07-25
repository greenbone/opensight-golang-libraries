// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package paging

// ValidateAndApplyPagingRules performs a validation of the original request and adds correct the correct values (defaults) if needed
func ValidateAndApplyPagingRules(model PagingSettingsInterface, request *Request) (*Request, error) {
	if err := validatePagingRequest(request); err != nil {
		// If there is an error in the request for Paging, get the defaults and set them on the request
		rowSize := model.GetPagingDefault()
		if request == nil {
			request = &Request{}
		}
		request.PageIndex = 0
		request.PageSize = rowSize
	}
	return request, nil
}
