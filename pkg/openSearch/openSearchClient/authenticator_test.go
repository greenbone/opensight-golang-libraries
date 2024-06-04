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
			name: "returns none auth method when configured",
			config: config.OpensearchClientConfig{
				AuthMethod: "none",
			},
			expectedMethod: none,
			expectedError:  false,
		},
		{
			name: "returns basic auth method when configured",
			config: config.OpensearchClientConfig{
				AuthMethod: "basic",
				Username:   "username",
				Password:   "password",
			},
			expectedMethod: basic,
			expectedError:  false,
		},
		{
			name: "returns error when username is empty for basic auth",
			config: config.OpensearchClientConfig{
				AuthMethod: "basic",
				Username:   "",
				Password:   "password",
			},
			expectedMethod: "",
			expectedError:  true,
		},
		{
			name: "returns openid auth method when configured",
			config: config.OpensearchClientConfig{
				AuthMethod:      "openid",
				IDPClientID:     "clientID",
				IDPClientSecret: "clientSecret",
			},
			expectedMethod: openId,
			expectedError:  false,
		},
		{
			name: "returns error when client secret is empty for openID auth",
			config: config.OpensearchClientConfig{
				AuthMethod:      "openid",
				IDPClientID:     "clientID",
				IDPClientSecret: "",
			},
			expectedMethod: "",
			expectedError:  true,
		},
		{
			name: "returns error when invalid auth method is configured",
			config: config.OpensearchClientConfig{
				AuthMethod: "",
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
	}{
		{
			name:           "injects basic auth header when basic auth method is used",
			authMethod:     basic,
			expectedHeader: "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
		},
		{
			name:           "injects bearer token when openid auth method is used",
			authMethod:     openId,
			expectedHeader: "Bearer mockToken",
		},
		{
			name:           "does not inject auth header when none auth method is used",
			authMethod:     none,
			expectedHeader: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authenticator := &Authenticator{
				config: config.OpensearchClientConfig{
					Username:        "username",
					Password:        "password",
					IDPClientID:     "clientID",
					IDPClientSecret: "clientSecret",
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
