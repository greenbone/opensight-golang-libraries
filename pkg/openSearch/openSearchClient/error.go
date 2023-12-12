// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import "github.com/pkg/errors"

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

func NewOpenSearchErrorWithStack(message string) error {
	return errors.WithStack(NewOpenSearchError(message))
}

// OpenSearchResourceAlreadyExists openSearch resource already exists
type OpenSearchResourceAlreadyExists struct {
	Message string
}

func (o *OpenSearchResourceAlreadyExists) Error() string {
	return o.Message
}

func NewOpenSearchResourceAlreadyExists(message string) *OpenSearchResourceAlreadyExists {
	return &OpenSearchResourceAlreadyExists{
		Message: message,
	}
}

func NewOpenSearchResourceAlreadyExistsWithStack(message string) error {
	return errors.WithStack(NewOpenSearchResourceAlreadyExists(message))
}

// OpenSearchResourceNotFound openSearch resource already exists
type OpenSearchResourceNotFound struct {
	Message string
}

func (o *OpenSearchResourceNotFound) Error() string {
	return o.Message
}

func NewOpenSearchResourceNotFound(message string) *OpenSearchResourceNotFound {
	return &OpenSearchResourceNotFound{
		Message: message,
	}
}

func NewOpenSearchResourceNotFoundWithStack(message string) error {
	return errors.WithStack(NewOpenSearchResourceNotFound(message))
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
