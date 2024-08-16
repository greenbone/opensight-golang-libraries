// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package notifications

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateNotification(t *testing.T) {
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
			name:         "notification can be sent",
			serverErrors: serverErrors{}, // no server errors
			wantErr:      false,
		},
		{
			name:         "notification can be sent despite temporary server failure",
			serverErrors: serverErrors{retryableFail: true},
			wantErr:      false,
		},
		{
			name:         "client returns error on non retryable notification service error",
			serverErrors: serverErrors{fatalFail: true},
			wantErr:      true,
		},
	}

	notification := Notification{
		Origin:       "Example Task XY",
		Timestamp:    time.Time{},
		Title:        "Example Task XY failed",
		Detail:       "Example Task XY failed because ...",
		Level:        LevelError,
		CustomFields: map[string]any{"extraProperty": "value"},
	}

	config := Config{
		Address:      "", // set below in test
		MaxRetries:   1,
		RetryWaitMin: time.Microsecond, // keep test short
		RetryWaitMax: time.Second,
	}

	wantNotificationSerialized := `{
		"origin": "Example Task XY",
		"timestamp": "0001-01-01T00:00:00Z",
		"title": "Example Task XY failed",
		"detail": "Example Task XY failed because ...",
		"level": "error",
		"customFields": {
			"extraProperty": "value"
		}
	}`

	wantRequestUri := basePath + createNotificationPath

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverCallCount atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Log("server called")
				serverCallCount.Add(1)

				if !assert.Equal(t, wantRequestUri, r.RequestURI, "invalid request url") {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// check body
				requestBody, err := io.ReadAll(r.Body)
				require.NoError(t, err, "failed to read request body")
				assert.JSONEq(t, string(wantNotificationSerialized),
					string(requestBody), "request body is not set properly")

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

				w.WriteHeader(http.StatusCreated)
			}))
			defer server.Close()

			config.Address = server.URL
			client := NewClient(http.DefaultClient, config)
			err := client.CreateNotification(context.Background(), notification)

			require.True(t, serverCallCount.Load() > 0, "server was not called")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
