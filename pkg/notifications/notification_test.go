// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package notifications

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type serverErrors struct { // set at most one of the fields to true
	fatalFail          bool
	retryableFail      bool
	authenticationFail bool
	unexpectedResponse bool
}

func TestClient_CreateNotification(t *testing.T) {
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
		{
			name:         "client fails on authentication error",
			serverErrors: serverErrors{authenticationFail: true},
			wantErr:      true,
		},
		{
			name:         "sending notification fails after maximum number of retries",
			serverErrors: serverErrors{unexpectedResponse: true},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverCallCount atomic.Int32

			mockAuthServer := setupMockAuthServer(t, tt.serverErrors.authenticationFail)
			defer mockAuthServer.Close()

			mockNotificationServer := setupMockNotificationServer(t, &serverCallCount, tt.serverErrors)
			defer mockNotificationServer.Close()

			config := Config{
				Address:      mockNotificationServer.URL, // set below in test
				MaxRetries:   1,
				RetryWaitMin: time.Microsecond, // keep test short
				RetryWaitMax: time.Second,
			}

			authentication := KeycloakAuthentication{
				ClientID: "client_id",
				Username: "username",
				Password: "password",
				AuthURL:  mockAuthServer.URL,
			}

			client := NewClient(http.DefaultClient, config, authentication)
			err := client.CreateNotification(context.Background(), notification)

			if !tt.serverErrors.authenticationFail {
				require.True(t, serverCallCount.Load() > 0, "server was not called")
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func setupMockAuthServer(t *testing.T, failAuth bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failAuth {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		err := json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
		require.NoError(t, err)
	}))
}

func setupMockNotificationServer(t *testing.T, serverCallCount *atomic.Int32, errors serverErrors) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCallCount.Add(1)

		token := r.Header.Get("Authorization")
		require.Equal(t, "Bearer mock-token", token, "missing or incorrect Authorization header")

		if !assert.Equal(t, basePath+createNotificationPath, r.RequestURI, "invalid request url") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		requestBody, err := io.ReadAll(r.Body)
		require.NoError(t, err, "failed to read request body")
		assert.JSONEq(t, `{
			"origin": "Example Task XY",
			"timestamp": "0001-01-01T00:00:00Z",
			"title": "Example Task XY failed",
			"detail": "Example Task XY failed because ...",
			"level": "error",
			"customFields": {"extraProperty": "value"}
		}`, string(requestBody), "request body is incorrect")

		// error simulation (fail on first attempt, if configured)
		if serverCallCount.Load() == 1 {
			if errors.fatalFail {
				t.Log("fatal server fail as per test config")
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else if errors.retryableFail {
				t.Log("retryable server fail as per test config")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
		}
		if errors.unexpectedResponse {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}))
}
