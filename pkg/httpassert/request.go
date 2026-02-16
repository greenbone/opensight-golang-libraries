// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package httpassert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint:interfacebloat
// Request interface provides fluent HTTP request building.
type Request interface {
	Get(path string) Request
	Getf(format string, a ...interface{}) Request
	Post(path string) Request
	Postf(format string, a ...interface{}) Request
	Put(path string) Request
	Putf(format string, a ...interface{}) Request
	Delete(path string) Request
	Deletef(format string, a ...interface{}) Request
	Options(path string) Request
	Optionsf(format string, a ...interface{}) Request
	Patch(path string) Request
	Patchf(format string, a ...interface{}) Request

	Perform(verb string, path string) Request
	Performf(verb string, path string, a ...interface{}) Request

	AuthHeader(header string) Request
	Headers(headers map[string]string) Request
	Header(key, value string) Request
	AuthJwt(jwt string) Request

	ContentType(string) Request

	Content(string) Request
	JsonContent(string) Request
	JsonContentObject(any) Request
	ContentFile(string) Request
	JsonContentFile(path string) Request

	Expect() Response
	ExpectEventually(check func(r Response), timeout time.Duration, interval time.Duration) Response
}

type request struct {
	t *testing.T

	router http.Handler

	method      string
	url         string
	contentType string
	body        string
	headers     map[string]string

	response *httptest.ResponseRecorder
}

func (m *request) Perform(verb string, path string) Request {
	switch verb {
	case "Post", http.MethodPost:
		return m.Post(path)
	case "Put", http.MethodPut:
		return m.Put(path)
	case "Get", http.MethodGet:
		return m.Get(path)
	case "Delete", http.MethodDelete:
		return m.Delete(path)
	case "Options", http.MethodOptions:
		return m.Options(path)
	case "Patch", http.MethodPatch:
		return m.Patch(path)
	default:
		m.t.Fatalf("unknown verb: %s", verb)
		return m
	}
}

func (m *request) Performf(verb string, path string, a ...interface{}) Request {
	switch verb {
	case "Post", http.MethodPost:
		return m.Postf(path, a...)
	case "Put", http.MethodPut:
		return m.Putf(path, a...)
	case "Get", http.MethodGet:
		return m.Getf(path, a...)
	case "Delete", http.MethodDelete:
		return m.Deletef(path, a...)
	case "Options", http.MethodOptions:
		return m.Optionsf(path, a...)
	case "Patch", http.MethodPatch:
		return m.Patchf(path, a...)
	default:
		m.t.Fatalf("unknown verb: %s", verb)
		return m
	}
}

// responseImpl implements Response.
type responseImpl struct {
	t        *testing.T
	response *httptest.ResponseRecorder
	request  *request
}

// New returns a new Request instance for the given router.
func New(t *testing.T, router http.Handler) Request {
	return &request{
		t:       t,
		router:  router,
		headers: map[string]string{},
	}
}

func (m *request) Post(path string) Request {
	m.method = http.MethodPost
	m.url = path
	return m
}

func (m *request) Postf(format string, a ...interface{}) Request {
	m.method = http.MethodPost
	m.url = fmt.Sprintf(format, a...)
	return m
}

func (m *request) Put(path string) Request {
	m.method = http.MethodPut
	m.url = path
	return m
}

func (m *request) Putf(format string, a ...interface{}) Request {
	m.method = http.MethodPut
	m.url = fmt.Sprintf(format, a...)
	return m
}

func (m *request) Get(path string) Request {
	m.method = http.MethodGet
	m.url = path
	return m
}

func (m *request) Getf(format string, a ...interface{}) Request {
	m.method = http.MethodGet
	m.url = fmt.Sprintf(format, a...)
	return m
}

func (m *request) Options(path string) Request {
	m.method = http.MethodOptions
	m.url = path
	return m
}

func (m *request) Optionsf(format string, a ...interface{}) Request {
	m.method = http.MethodOptions
	m.url = fmt.Sprintf(format, a...)
	return m
}

func (m *request) Delete(path string) Request {
	m.method = http.MethodDelete
	m.url = path
	return m
}

func (m *request) Deletef(format string, a ...interface{}) Request {
	m.method = http.MethodDelete
	m.url = fmt.Sprintf(format, a...)
	return m
}

func (m *request) Patch(path string) Request {
	m.method = http.MethodPatch
	m.url = path
	return m
}

func (m *request) Patchf(format string, a ...interface{}) Request {
	m.method = http.MethodPatch
	m.url = fmt.Sprintf(format, a...)
	return m
}

func (m *request) AuthHeader(header string) Request {
	m.headers["Authorization"] = header
	return m
}

func (m *request) Headers(headers map[string]string) Request {
	m.headers = headers
	return m
}

func (m *request) Header(key, value string) Request {
	m.headers[key] = value
	return m
}

func (m *request) AuthJwt(jwt string) Request {
	m.AuthHeader("bearer " + jwt)
	return m
}

func (m *request) ContentType(ct string) Request {
	m.contentType = ct
	return m
}

func (m *request) Content(body string) Request {
	m.body = body
	return m
}

func (m *request) JsonContent(body string) Request {
	m.ContentType("application/json")
	m.Content(body)
	return m
}

func (m *request) JsonContentObject(obj any) Request {
	marshal, err := json.Marshal(obj)
	require.NoError(m.t, err)

	m.JsonContent(string(marshal))
	return m
}

func (m *request) ContentFile(path string) Request {
	content, err := os.ReadFile(path)
	if err != nil {
		assert.Fail(m.t, err.Error())
	}
	m.body = string(content)
	return m
}

func (m *request) JsonContentFile(path string) Request {
	m.ContentType("application/json")
	m.ContentFile(path)
	return m
}

func (m *request) Expect() Response {
	if m.router == nil {
		assert.Fail(m.t, "Request router is nil")
	}

	req, err := http.NewRequest(m.method, m.url, bytes.NewBufferString(m.body))
	if err != nil {
		assert.Fail(m.t, err.Error())
	}

	if m.contentType != "" {
		req.Header.Set("Content-Type", m.contentType)
	}

	for k, v := range m.headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	m.router.ServeHTTP(rec, req)
	m.response = rec

	return &responseImpl{t: m.t, response: rec, request: m}
}

func (m *request) ExpectEventually(assertions func(resp Response), timeout, interval time.Duration) Response {
	if m.router == nil {
		assert.Fail(m.t, "Request router is nil")
	}

	deadline := time.Now().Add(timeout)
	var lastResp *responseImpl

	for {
		req, err := http.NewRequest(m.method, m.url, bytes.NewBufferString(m.body))
		if err != nil {
			assert.Fail(m.t, err.Error())
			break
		}

		if m.contentType != "" {
			req.Header.Set("Content-Type", m.contentType)
		}

		for k, v := range m.headers {
			req.Header.Set(k, v)
		}

		rec := httptest.NewRecorder()
		m.router.ServeHTTP(rec, req)
		m.response = rec

		resp := &responseImpl{t: m.t, response: rec, request: m}
		lastResp = resp

		sandbox := &testing.T{}
		assertions(&responseImpl{t: sandbox, response: rec, request: m})

		if sandbox.Failed() {
			resp.Log()
		}

		if !sandbox.Failed() {
			return resp
		}

		if time.Now().After(deadline) {
			assert.Fail(m.t, fmt.Sprintf("ExpectEventually: condition not met within %v", timeout))
			resp.Log()
			return resp
		}

		time.Sleep(interval)
	}

	return lastResp
}
