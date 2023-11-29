// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_client

import (
	"bytes"
	"io"
	"net/http"
)

type MockTransport struct {
	Request     string
	Response    *http.Response
	RoundTripFn func(req *http.Request) (*http.Response, error)
}

func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var buf bytes.Buffer
	if req.Body != nil {
		tee := io.TeeReader(req.Body, &buf)

		value, err := io.ReadAll(tee)
		if err != nil {
			return nil, err
		}
		t.Request = string(value)
	}

	if t.RoundTripFn == nil {
		t.RoundTripFn = func(req *http.Request) (*http.Response, error) {
			return t.Response, nil
		}
	}

	response, err := t.RoundTripFn(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
