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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// nolint:interfacebloat
// RequestStart provides fluent HTTP *method + path* selection.
// Each call returns a fresh Request
type RequestStart interface {
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
}

// nolint:interfacebloat
// Request provides fluent request configuration
type Request interface {
	AuthHeader(header string) Request
	Headers(headers map[string]string) Request
	Header(key, value string) Request
	AuthJwt(jwt string) Request

	ContentType(string) Request

	Content(string) Request
	ContentFile(string) Request

	JsonContent(string) Request
	JsonContentTemplate(body string, values map[string]any) Request
	JsonContentObject(any) Request
	JsonContentFile(path string) Request

	Expect() Response
	ExpectEventually(check func(r Response), timeout time.Duration, interval time.Duration) Response
}

type starter struct {
	t      *testing.T
	router http.Handler
}

// request holds the per-request state (fresh instance per verb selection).
type request struct {
	t      *testing.T
	router http.Handler

	method      string
	url         string
	contentType string
	body        string
	headers     map[string]string

	response *httptest.ResponseRecorder
}

// responseImpl implements Response.
type responseImpl struct {
	t        *testing.T
	response *httptest.ResponseRecorder
	request  *request
}

// New returns a new RequestStart instance for the given router.
// All method calls (Get/Post/...) return a *fresh* Request.
func New(t *testing.T, router http.Handler) RequestStart {
	return &starter{t: t, router: router}
}

func (s *starter) newRequest(method, url string) Request {
	return &request{
		t:       s.t,
		router:  s.router,
		method:  method,
		url:     url,
		headers: map[string]string{}, // fresh map: no tinted headers
	}
}

func (s *starter) Post(path string) Request {
	return s.newRequest(http.MethodPost, path)
}

func (s *starter) Postf(format string, a ...interface{}) Request {
	return s.newRequest(http.MethodPost, fmt.Sprintf(format, a...))
}

func (s *starter) Put(path string) Request {
	return s.newRequest(http.MethodPut, path)
}

func (s *starter) Putf(format string, a ...interface{}) Request {
	return s.newRequest(http.MethodPut, fmt.Sprintf(format, a...))
}

func (s *starter) Get(path string) Request {
	return s.newRequest(http.MethodGet, path)
}

func (s *starter) Getf(format string, a ...interface{}) Request {
	return s.newRequest(http.MethodGet, fmt.Sprintf(format, a...))
}

func (s *starter) Options(path string) Request {
	return s.newRequest(http.MethodOptions, path)
}
func (s *starter) Optionsf(format string, a ...interface{}) Request {
	return s.newRequest(http.MethodOptions, fmt.Sprintf(format, a...))
}

func (s *starter) Delete(path string) Request {
	return s.newRequest(http.MethodDelete, path)
}

func (s *starter) Deletef(format string, a ...interface{}) Request {
	return s.newRequest(http.MethodDelete, fmt.Sprintf(format, a...))
}

func (s *starter) Patch(path string) Request {
	return s.newRequest(http.MethodPatch, path)
}

func (s *starter) Patchf(format string, a ...interface{}) Request {
	return s.newRequest(http.MethodPatch, fmt.Sprintf(format, a...))
}

func (s *starter) Perform(verb string, path string) Request {
	switch verb {
	case "Post", http.MethodPost:
		return s.Post(path)
	case "Put", http.MethodPut:
		return s.Put(path)
	case "Get", http.MethodGet:
		return s.Get(path)
	case "Delete", http.MethodDelete:
		return s.Delete(path)
	case "Options", http.MethodOptions:
		return s.Options(path)
	case "Patch", http.MethodPatch:
		return s.Patch(path)
	default:
		s.t.Fatalf("unknown verb: %s", verb)
		return &request{} // unreachable, but keeps compiler happy
	}
}

func (s *starter) Performf(verb string, format string, a ...interface{}) Request {
	switch verb {
	case "Post", http.MethodPost:
		return s.Postf(format, a...)
	case "Put", http.MethodPut:
		return s.Putf(format, a...)
	case "Get", http.MethodGet:
		return s.Getf(format, a...)
	case "Delete", http.MethodDelete:
		return s.Deletef(format, a...)
	case "Options", http.MethodOptions:
		return s.Optionsf(format, a...)
	case "Patch", http.MethodPatch:
		return s.Patchf(format, a...)
	default:
		s.t.Fatalf("unknown verb: %s", verb)
		return &request{} // unreachable, but keeps compiler happy
	}
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

func (m *request) ContentFile(path string) Request {
	content, err := os.ReadFile(path)
	if err != nil {
		assert.Fail(m.t, err.Error())
	}
	m.body = string(content)
	return m
}

func (m *request) JsonContent(body string) Request {
	m.ContentType("application/json")
	m.Content(body)
	return m
}

func (m *request) JsonContentTemplate(body string, values map[string]any) Request {
	m.ContentType("application/json")

	jsonBody := body
	// apply provided values into the template
	for k, v := range values {
		// normalize JSONPath-like keys (convert $.a[0].b to a.0.b)
		key := strings.TrimPrefix(k, "$.")
		key = strings.ReplaceAll(key, "[", ".")
		key = strings.ReplaceAll(key, "]", "")

		if !gjson.Get(jsonBody, key).Exists() {
			assert.Fail(m.t, "Json key does not exist in template: "+k)
		}

		tmp, err := sjson.Set(jsonBody, key, v)
		if err != nil {
			assert.Fail(m.t, "JsonTemplate set value failed: "+err.Error())
			return m
		}
		jsonBody = tmp
	}

	m.Content(jsonBody)
	return m
}

func (m *request) JsonContentObject(obj any) Request {
	marshal, err := json.Marshal(obj)
	require.NoError(m.t, err)

	m.JsonContent(string(marshal))
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
