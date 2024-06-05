// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"net/http"
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/stretchr/testify/assert"
)

type MockTokenReceiver struct{}

func (m *MockTokenReceiver) GetClientAccessToken(clientName, clientSecret string) (string, error) {
	return "mockToken", nil
}

func TestGetAuthenticationMethod(t *testing.T) {
	mockTokenReceiver := &MockTokenReceiver{}

	tests := []struct {
		name           string
		config         config.OpensearchClientConfig
		expectedMethod authMethod
		expectedError  bool
	}{
		{
			name: "returns basic auth method when configured",
			config: config.OpensearchClientConfig{
				AuthMethod:   "basic",
				AuthUsername: "username",
				AuthPassword: "password",
			},
			expectedMethod: basic,
			expectedError:  false,
		},
		{
			name: "returns openid auth method when configured",
			config: config.OpensearchClientConfig{
				AuthMethod:   "openid",
				AuthUsername: "clientID",
				AuthPassword: "clientSecret",
			},
			expectedMethod: openId,
			expectedError:  false,
		},
		{
			name: "returns error when password is empty",
			config: config.OpensearchClientConfig{
				AuthMethod: "basic",
				AuthUsername: "username",
				AuthPassword: "",
			},
			expectedMethod: "",
			expectedError:  true,
		},
		{
			name: "returns error when invalid auth method is configured",
			config: config.OpensearchClientConfig{
				AuthMethod: "",
				AuthUsername: "username",
				AuthPassword: "password",
			},
			expectedMethod: "",
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			method, err := getAuthenticationMethod(tc.config, mockTokenReceiver)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMethod, method)
			}
		})
	}
}

func TestInjectAuthenticationHeader(t *testing.T) {
	mockTokenReceiver := &MockTokenReceiver{}

	tests := []struct {
		name           string
		authMethod     authMethod
		expectedHeader string
		authUser       string
		authPass       string
	}{
		{
			name:           "injects basic auth header when basic auth method is used",
			authMethod:     basic,
			authUser:       "username",
			authPass:       "password",
			expectedHeader: "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
		},
		{
			name:           "injects bearer token when openid auth method is used",
			authMethod:     openId,
			authUser:       "clientID",
			authPass:       "clientSecret",
			expectedHeader: "Bearer mockToken",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authenticator := &Authenticator{
				config: config.OpensearchClientConfig{
					AuthUsername: tc.authUser,
					AuthPassword: tc.authPass,
				},
				tokenReceiver: mockTokenReceiver,
				authMethod:    tc.authMethod,
			}

			req, _ := http.NewRequest("GET", "http://localhost", nil)
			reqWithAuth := authenticator.injectAuthenticationHeader(authenticator.authMethod, req)

			assert.Equal(t, tc.expectedHeader, reqWithAuth.Header.Get("Authorization"))
		})
	}
}
