// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package openSearchClient provides functionality for interacting with OpenSearch.
package openSearchClient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchtransport"
	"github.com/rs/zerolog/log"
)

// authMethod represents the type of authentication method used.
type authMethod string

// Constants representing the different types of authentication methods.
const (
	basic  authMethod = "basic"
	openId authMethod = "openid"
)

// TokenReceiver is an interface for receiving client access tokens.
type TokenReceiver interface {
	GetClientAccessToken(clientName, clientSecret string) (string, error)
	ClearClientAccessToken()
}

// Authenticator is a struct that holds the necessary information for authenticating with OpenSearch.
type Authenticator struct {
	clientTransport opensearchtransport.Interface
	config          config.OpensearchClientConfig
	tokenReceiver   TokenReceiver
	authMethod      authMethod
}

// Ensure Authenticator implements the opensearchtransport.Interface interface.
var _ opensearchtransport.Interface = &Authenticator{}

// InjectAuthenticationIntoClient is a function that sets up the authentication method for the OpenSearch client.
// client is the OpenSearch client to inject the authentication into.
// config is the configuration for the OpenSearch client.
// tokenReceiver is the token receiver for OpenID authentication and must implement the GetClientAccessToken function. It can be nil for basic authentication.
func InjectAuthenticationIntoClient(client *opensearchapi.Client,
	config config.OpensearchClientConfig, tokenReceiver TokenReceiver,
) error {
	method, err := getAuthenticationMethod(config, tokenReceiver)
	if err != nil {
		return err
	}
	log.Debug().
		Str("opensearch_auth_method", string(method)).
		Msgf("set up auth method for opensearch client: %s", method)

	authenticator := &Authenticator{
		config:        config,
		tokenReceiver: tokenReceiver,
		authMethod:    method,
	}

	// store the original Transport interface implementation of the opensearch client to be able to wrap it
	// in the authenticator.Perform method
	authenticator.clientTransport = client.Client.Transport
	// replace the original Transport interface implementation of the opensearch client with the wrapper
	client.Client.Transport = authenticator

	return nil
}

// getAuthenticationMethod is a helper function that determines the authentication method based on the provided configuration.
func getAuthenticationMethod(conf config.OpensearchClientConfig, tokenReceiver TokenReceiver) (authMethod, error) {
	if conf.AuthUsername == "" || conf.AuthPassword == "" {
		return "", fmt.Errorf("username and password must be set in configuration")
	}

	method := authMethod(strings.ToLower(conf.AuthMethod))

	switch method {
	case basic:
	case openId:
		if tokenReceiver == nil {
			return "", fmt.Errorf("token receiver must not be nil for openid authentication")
		}
	default:
		return "", fmt.Errorf("invalid authentication method for opensearch: %s", conf.AuthMethod)
	}

	return method, nil
}

// injectAuthenticationHeader is a method that injects the appropriate authentication header into the request.
func (a *Authenticator) injectAuthenticationHeader(req *http.Request) (*http.Request, error) {
	reqClone := req.Clone(req.Context())
	if reqClone.Header == nil {
		reqClone.Header = http.Header{}
	}
	switch a.authMethod {
	case basic:
		reqClone.SetBasicAuth(a.config.AuthUsername, a.config.AuthPassword)
	case openId:
		token, err := a.tokenReceiver.GetClientAccessToken(a.config.AuthUsername, a.config.AuthPassword)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve authorization header: %w", err)
		}
		reqClone.Header.Set("Authorization", "Bearer "+token)
	default:
		return nil, fmt.Errorf("undefined authentication method for opensearch client: %s", a.authMethod)
	}
	return reqClone, nil
}

func (a *Authenticator) handleUnauthorized() {
	if a.authMethod == openId {
		a.tokenReceiver.ClearClientAccessToken()
	}
}

// Perform is a method that implements the opensearchtransport.Interface interface.
// It injects the authentication header into the request and then performs the request.
func (a *Authenticator) Perform(req *http.Request) (*http.Response, error) {
	requestWithInjectedAuth, err := a.injectAuthenticationHeader(req)
	if err != nil {
		return nil, err
	}

	resp, err := a.clientTransport.Perform(requestWithInjectedAuth)
	if resp != nil && resp.StatusCode >= http.StatusBadRequest &&
		resp.StatusCode < http.StatusInternalServerError {
		a.handleUnauthorized()
	}

	return resp, err
}
