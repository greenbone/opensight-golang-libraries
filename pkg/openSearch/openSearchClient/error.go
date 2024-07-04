// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

type OpenSearchErrors struct {
	Reasons []OpenSearchRootCause
	Type    string
	Reason  string
}

type OpenSearchRootCause struct {
	Type   string
	Reason string
}

type OpenSearchErrorResponse struct {
	Error  OpenSearchErrors
	Status int
}

// OpenSearchError openSearch error
type OpenSearchError struct {
	Message string
}

func (o *OpenSearchError) Error() string {
	return o.Message
}

func NewOpenSearchError(message string) *OpenSearchError {
	return &OpenSearchError{
		Message: message,
	}
}

// OpenSearchResourceAlreadyExists openSearch resource already exists
type OpenSearchResourceAlreadyExists struct {
	Message string
}

func (o *OpenSearchResourceAlreadyExists) Error() string {
	return o.Message
}

// OpenSearchResourceNotFound openSearch resource already exists
type OpenSearchResourceNotFound struct {
	Message string
}

func (o *OpenSearchResourceNotFound) Error() string {
	return o.Message
}

type IndexError struct {
	Index DocumentError `json:"index"`
}
type DocumentError struct {
	IndexName  string            `json:"_index"`
	IndexType  string            `json:"_type"`
	DocumentId string            `json:"_id"`
	StatusCode uint              `json:"status"`
	Error      DocumentErrorType `json:"error"`
}

type DocumentErrorType struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}
