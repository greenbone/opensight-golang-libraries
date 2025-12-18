// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeycloakClient_GetToken(t *testing.T) {
	tests := map[string]struct {
		responseBody string
		responseCode int
		wantErr      bool
		wantToken    string
	}{
		"successful token retrieval": {
			responseBody: `{"access_token": "test-token", "expires_in": 3600}`,
			responseCode: http.StatusOK,
			wantErr:      false,
			wantToken:    "test-token",
		},
		"failed authentication": {
			responseBody: `{}`,
			responseCode: http.StatusUnauthorized,
			wantErr:      true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var serverCallCount atomic.Int32
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverCallCount.Add(1)
				w.WriteHeader(tt.responseCode)
				_, err := w.Write([]byte(tt.responseBody))
				require.NoError(t, err)
			}))
			defer mockServer.Close()

			client := NewKeycloakClient(http.DefaultClient, KeycloakConfig{
				AuthURL: mockServer.URL,
				// the other fields are also required in real scenario, but omit here for brevity
			})
			gotToken, err := client.GetToken(context.Background())
			assert.True(t, serverCallCount.Load() > 0, "server was not called")

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, gotToken)
			}
		})
	}
}

type fakeClock struct {
	currentTime time.Time
}

func (fc *fakeClock) Now() time.Time {
	return fc.currentTime
}

func (fc *fakeClock) Advance(d time.Duration) {
	fc.currentTime = fc.currentTime.Add(d)
}

func NewFakeClock(startTime time.Time) *fakeClock {
	return &fakeClock{currentTime: startTime}
}

func TestKeycloakClient_GetToken_Refresh(t *testing.T) {
	tokenValidity := 60 * time.Second
	kcMockResponse := []byte(fmt.Sprintf(`{"access_token": "test-token", "expires_in": %d}`, int(tokenValidity.Seconds())))

	tests := map[string]struct {
		responseBody     string
		responseCode     int
		requestAfter     time.Duration
		wantServerCalled int
		wantToken        string
	}{
		"token is cached": {
			requestAfter:     tokenValidity - tokenRefreshMargin - time.Nanosecond,
			wantServerCalled: 1, // should be called only once due to caching
			wantToken:        "test-token",
		},
		"token expiry handling": {
			requestAfter:     tokenValidity - tokenRefreshMargin + time.Nanosecond,
			wantServerCalled: 2, // should be called twice due to expiry
			wantToken:        "test-token",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fakeClock := NewFakeClock(time.Now())

			var serverCallCount atomic.Int32
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverCallCount.Add(1)
				w.WriteHeader(200)
				_, err := w.Write(kcMockResponse)
				require.NoError(t, err)
			}))
			defer mockServer.Close()

			client := NewKeycloakClient(http.DefaultClient, KeycloakConfig{
				AuthURL: mockServer.URL,
				// the other fields are also required in real scenario, but omit here for brevity
			})
			client.clock = fakeClock

			_, err := client.GetToken(context.Background())
			require.NoError(t, err)

			fakeClock.Advance(tc.requestAfter)

			gotToken, err := client.GetToken(context.Background()) // second call to test caching/refresh
			require.NoError(t, err)

			assert.Equal(t, tc.wantServerCalled, int(serverCallCount.Load()), "unexpected number of server calls")
			assert.Equal(t, tc.wantToken, gotToken)
		})
	}
}
