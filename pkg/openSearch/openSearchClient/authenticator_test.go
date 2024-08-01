// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTokenReceiver struct {
	mock.Mock
}

func (m *MockTokenReceiver) GetClientAccessToken(clientName, clientSecret string) (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestGetAuthenticationMethod(t *testing.T) {
	tests := []struct {
		name           string
		config         config.OpensearchClientConfig
		tokenReceiver  TokenReceiver
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
			tokenReceiver:  nil,
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
			tokenReceiver:  &MockTokenReceiver{},
			expectedMethod: openId,
			expectedError:  false,
		},
		{
			name: "returns error when username is empty",
			config: config.OpensearchClientConfig{
				AuthMethod:   "basic",
				AuthUsername: "",
				AuthPassword: "password",
			},
			expectedMethod: "",
			expectedError:  true,
		},
		{
			name: "returns error when password is empty",
			config: config.OpensearchClientConfig{
				AuthMethod:   "basic",
				AuthUsername: "username",
				AuthPassword: "",
			},
			expectedMethod: "",
			expectedError:  true,
		},
		{
			name: "returns error when token receiver is nil for openid auth method",
			config: config.OpensearchClientConfig{
				AuthMethod:   "openid",
				AuthUsername: "username",
				AuthPassword: "password",
			},
			tokenReceiver:  nil,
			expectedMethod: "",
			expectedError:  true,
		},
		{
			name: "returns error when invalid auth method is configured",
			config: config.OpensearchClientConfig{
				AuthMethod:   "",
				AuthUsername: "username",
				AuthPassword: "password",
			},
			expectedMethod: "",
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			method, err := getAuthenticationMethod(tc.config, tc.tokenReceiver)

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
	tests := []struct {
		name           string
		authMethod     authMethod
		authUser       string
		authPass       string
		mockToken      string
		mockError      error
		expectedHeader string
		expectedError  bool
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
			mockToken:      "mockToken",
			mockError:      nil,
			expectedHeader: "Bearer mockToken",
		},
		{
			name:           "returns error when tokenReceiver.GetClientAccessToken fails",
			authMethod:     openId,
			authUser:       "clientID",
			authPass:       "clientSecret",
			mockToken:      "",
			mockError:      fmt.Errorf("mock error"),
			expectedHeader: "",
			expectedError:  true,
		},
		{
			name:          "returns error when undefined auth method is configured",
			authMethod:    authMethod("undefined"),
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTokenReceiver := new(MockTokenReceiver)
			mockTokenReceiver.On("GetClientAccessToken").Return(tc.mockToken, tc.mockError)

			authenticator := &Authenticator{
				config: config.OpensearchClientConfig{
					AuthUsername: tc.authUser,
					AuthPassword: tc.authPass,
				},
				tokenReceiver: mockTokenReceiver,
				authMethod:    tc.authMethod,
			}

			req, _ := http.NewRequest("GET", "http://localhost", nil)
			reqWithAuth, err := authenticator.injectAuthenticationHeader(req)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedHeader, reqWithAuth.Header.Get("Authorization"))
			}
		})
	}
}

type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) Perform(req *http.Request) (*http.Response, error) {
	args := m.Called()
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestInjectAuthenticationIntoClient(t *testing.T) {
	mockTransport := new(MockTransport)
	mockTransport.On("Perform").Return(&http.Response{}, nil)

	client := &opensearchapi.Client{
		Client: &opensearch.Client{
			Transport: mockTransport,
		},
	}

	conf := config.OpensearchClientConfig{
		AuthMethod:   "openid",
		AuthUsername: "username",
		AuthPassword: "password",
	}

	mockTokenReceiver := new(MockTokenReceiver)
	mockTokenReceiver.On("GetClientAccessToken").Return("mockToken", nil)

	err := InjectAuthenticationIntoClient(client, conf, mockTokenReceiver)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "http://localhost", nil)

	_, err = client.Client.Perform(req)
	assert.NoError(t, err)

	mockTransport.AssertCalled(t, "Perform")
	mockTokenReceiver.AssertCalled(t, "GetClientAccessToken")
}
