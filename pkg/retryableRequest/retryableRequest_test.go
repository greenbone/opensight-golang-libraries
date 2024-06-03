// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package retryableRequest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ExecuteRequestWithRetry(t *testing.T) {
	type serverErrors struct { // set at most one of the fields to true
		fatalFail     bool
		retryableFail bool
	}
	tests := []struct {
		name         string
		serverErrors serverErrors
		wantErr      bool
	}{
		{
			name:         "request can be executed",
			serverErrors: serverErrors{}, // no server errors
			wantErr:      false,
		},
		{
			name:         "request can be executed despite temporary server failure",
			serverErrors: serverErrors{retryableFail: true},
			wantErr:      false,
		},
		{
			name:         "returns error on non retryable error",
			serverErrors: serverErrors{fatalFail: true},
			wantErr:      true,
		},
	}

	// input
	maxRetries := 5
	retryWaitMin := time.Microsecond // keep test short
	retryWaitMax := time.Second
	urlPath := "/something"
	bodyContent := `{ "key": "value" }`

	responseCode := http.StatusCreated
	responseContent := `{ "key": "response value" }`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverCallCount atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Log("server called")
				serverCallCount.Add(1)

				if !assert.Equal(t, urlPath, r.RequestURI, "invalid request url") {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// check body
				requestBody, err := io.ReadAll(r.Body)
				require.NoError(t, err, "failed to read request body")
				assert.JSONEq(t, bodyContent, string(requestBody), "request body is not set properly")

				// error simulation (fail on first attempt, if configured)
				if serverCallCount.Load() == 1 {
					if tt.serverErrors.fatalFail {
						t.Log("fatal server fail as per test config")
						w.WriteHeader(http.StatusUnauthorized)
						return
					} else if tt.serverErrors.retryableFail {
						t.Log("retryable server fail as per test config")
						w.WriteHeader(http.StatusTooManyRequests)
						return
					}
				}

				w.WriteHeader(responseCode)
				w.Write([]byte(responseContent))
			}))
			defer server.Close()

			req, err := http.NewRequest(http.MethodPost, server.URL+urlPath, strings.NewReader(bodyContent))
			require.NoError(t, err, "could not build request")

			response, err := ExecuteRequestWithRetry(context.Background(), http.DefaultClient, req, maxRetries, retryWaitMin, retryWaitMax)

			require.True(t, serverCallCount.Load() > 0, "server was not called")
			if tt.serverErrors.fatalFail {
				require.True(t, serverCallCount.Load() < int32(maxRetries+1), "did not stop retrying of fatal failure")
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, responseCode, response.StatusCode)
				body, err := io.ReadAll(response.Body)
				require.NoError(t, err, "could not read response body")

				assert.Equal(t, []byte(responseContent), body)
			}
		})
	}
}

func Test_ExecuteRequestWithRetry_Cancel(t *testing.T) {

	// config
	maxRetries := 5
	retryWaitMin := time.Microsecond // keep test short
	retryWaitMax := time.Second

	cancelInAttempt := maxRetries - 1

	ctx, cancel := context.WithCancel(context.Background())
	var serverCallCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("server called")
		serverCallCount.Add(1)

		if serverCallCount.Load() == int32(cancelInAttempt) {
			// of course in a real scenario `cancel` would be called somewhere in the client code ...
			cancel() // ... but calling it here keeps the test setup simple, as we know exactly at which request the client is supposed to stop
		}

		w.WriteHeader(http.StatusTooManyRequests)

	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
	require.NoError(t, err, "could not build request")

	_, err = ExecuteRequestWithRetry(ctx, http.DefaultClient, req, maxRetries, retryWaitMin, retryWaitMax)
	assert.True(t, errors.Is(err, context.Canceled), "we should return an error which indicates the cancellation")

	assert.Equal(t, serverCallCount.Load(), int32(cancelInAttempt), "there was a retry after cancellation or not enough retries were attempted")
}

func Test_DeepCopyRequest(t *testing.T) {
	bodyContent := `{ "key": "value" }`

	req, err := http.NewRequest(http.MethodPost, "https:/greenbone.net", bytes.NewReader([]byte(bodyContent)))
	req.Header.Set("SomeHeader", "SomeValue")
	require.NoError(t, err, "failed to build request")

	copy, err := DeepCopyRequest(req)
	require.NoError(t, err, "Deep copy should not fail")

	// compare body contents
	reqBody, err := io.ReadAll(req.Body)
	require.NoError(t, err, "failed to read body")
	assert.Equal(t, bodyContent, string(reqBody), "request does not contain expected body (was it not properly cloned and emptied by a previous read?)")

	copyBody, err := io.ReadAll(copy.Body)
	require.NoError(t, err, "failed to read body")
	assert.Equal(t, bodyContent, string(copyBody), "request copy does not contain expected body (was it not properly cloned and emptied by a previous read?)")

	req.Body = nil
	copy.Body = nil

	// `assert.Equal` can't compare functions, and this field is not relevant for our case
	req.GetBody = nil
	copy.GetBody = nil
	assert.Equal(t, req, copy, "copy is not an exact copy")
}
