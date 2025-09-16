// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package notifications

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serverErrors struct { // set at most one of the fields to true
	fatalFail          bool
	retryableFail      bool
	authenticationFail bool
	unexpectedResponse bool
}

const checkForCurrentTimestamp = "marker to check for current timestamp"

func TestClient_CreateNotification(t *testing.T) {
	notification := Notification{
		Origin:       "Example Task XY",
		Timestamp:    time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC),
		Title:        "Example Task XY failed",
		Detail:       "Example Task XY failed because ...",
		Level:        LevelError,
		CustomFields: map[string]any{"extraProperty": "value"},
	}

	notificationWithoutTimestamp := notification
	notificationWithoutTimestamp.Timestamp = time.Time{}

	wantNotification := notificationModel{
		Origin:       "Example Task XY",
		Timestamp:    "0001-02-03T04:05:06.000000007Z",
		Title:        "Example Task XY failed",
		Detail:       "Example Task XY failed because ...",
		Level:        LevelError,
		CustomFields: map[string]any{"extraProperty": "value"},
	}
	wantNotificationWithoutTimestamp := wantNotification
	wantNotificationWithoutTimestamp.Timestamp = checkForCurrentTimestamp // can't test exact timestamp

	tests := []struct {
		name                 string
		notification         Notification
		serverErrors         serverErrors
		wantNotificationSent notificationModel // expected on server side
		wantErr              bool
	}{
		{
			name:                 "notification can be sent",
			notification:         notification,
			serverErrors:         serverErrors{}, // no server errors
			wantNotificationSent: wantNotification,
			wantErr:              false,
		},
		{
			name:                 "success, adding timestamp when unset",
			notification:         notificationWithoutTimestamp,
			serverErrors:         serverErrors{}, // no server errors
			wantNotificationSent: wantNotificationWithoutTimestamp,
			wantErr:              false,
		},
		{
			name:                 "notification can be sent despite temporary server failure",
			notification:         notification,
			serverErrors:         serverErrors{retryableFail: true},
			wantNotificationSent: wantNotification,
			wantErr:              false,
		},
		{
			name:                 "client returns error on non retryable notification service error",
			notification:         notification,
			serverErrors:         serverErrors{fatalFail: true},
			wantNotificationSent: wantNotification,
			wantErr:              true,
		},
		{
			name:                 "client fails on authentication error",
			notification:         notification,
			serverErrors:         serverErrors{authenticationFail: true},
			wantNotificationSent: wantNotification,
			wantErr:              true,
		},
		{
			name:                 "sending notification fails after maximum number of retries",
			notification:         notification,
			serverErrors:         serverErrors{unexpectedResponse: true},
			wantNotificationSent: wantNotification,
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverCallCount atomic.Int32

			mockAuthServer := setupMockAuthServer(t, tt.serverErrors.authenticationFail)
			defer mockAuthServer.Close()

			mockNotificationServer := setupMockNotificationServer(t, &serverCallCount, tt.wantNotificationSent, tt.serverErrors)
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
			err := client.CreateNotification(context.Background(), tt.notification)

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

func setupMockNotificationServer(t *testing.T, serverCallCount *atomic.Int32, wantNotification notificationModel, errors serverErrors) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCallCount.Add(1)

		token := r.Header.Get("Authorization")
		require.Equal(t, "Bearer mock-token", token,
			"missing or incorrect Authorization header")

		if !assert.Equal(t, basePath+createNotificationPath, r.RequestURI, "invalid request url") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		requestBody, err := io.ReadAll(r.Body)
		require.NoError(t, err, "failed to read request body")
		var gotNotification notificationModel
		err = json.Unmarshal(requestBody, &gotNotification)
		require.NoError(t, err, "failed to unmarshal request body")

		if wantNotification.Timestamp == checkForCurrentTimestamp {
			gotTime, err := time.Parse(time.RFC3339Nano, gotNotification.Timestamp)
			assert.NoError(t, err, "failed to parse timestamp from request body")

			assert.NotEqual(t, time.Time{}, gotTime, "cliend did not set timestamp")
			assert.True(t, gotTime.Before(time.Now()), "timestamp is in the future")
			assert.True(t, gotTime.After(time.Now().Add(-time.Minute)), "timestamp is too far in the past")
			// timestamp was checked, ignore in other comparisons
			wantNotification.Timestamp = gotNotification.Timestamp
		}
		assert.Equal(t, wantNotification, gotNotification)

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
