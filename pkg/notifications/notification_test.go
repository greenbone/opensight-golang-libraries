// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package notifications

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateNotification(t *testing.T) {
	tests := []struct {
		name                   string
		serverHttpResponseCode int
		wantErr                bool
	}{
		{
			name:                   "notification can be sent",
			serverHttpResponseCode: http.StatusOK,
			wantErr:                false,
		},
		{
			name:                   "client returns error on notification service error",
			serverHttpResponseCode: http.StatusInternalServerError,
			wantErr:                true,
		},
	}

	// client arguments
	notification := Notification{
		Title: "test notification",
		// other mandatory fields not set for brevity ...
	}

	wantNotificationSerialized, err := json.Marshal(notification)
	require.NoError(t, err, "failed to parse notification to json")

	wantRequestUri := basePath + createNotificationPath

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverCalled atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Log("server called")
				serverCalled.Store(true)

				if !assert.Equal(t, wantRequestUri, r.RequestURI, "invalid request url") {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// check body
				requestBody, err := io.ReadAll(r.Body)
				require.NoError(t, err, "failed to read request body")
				assert.JSONEq(t, string(wantNotificationSerialized), string(requestBody), "request body is not set properly")

				w.WriteHeader(tt.serverHttpResponseCode)
			}))
			defer server.Close()

			notificationServiceBaseUrl := server.URL

			client := NewClient(http.DefaultClient, notificationServiceBaseUrl)
			err := client.CreateNotification(notification)

			require.True(t, serverCalled.Load(), "server was not called")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
