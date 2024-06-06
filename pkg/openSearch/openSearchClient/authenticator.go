// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package openSearchClient provides functionality for interacting with OpenSearch.
package openSearchClient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/rs/zerolog/log"
)

// authMethod represents the type of authentication method used.
type authMethod string

// Constants representing the different types of authentication methods.
const (
	basic  authMethod = "basic"
	openId authMethod = "openid"
)

// ITokenReceiver is an interface for receiving client access tokens.
type ITokenReceiver interface {
	GetClientAccessToken(clientName, clientSecret string) (string, error)
}

// Authenticator is a struct that holds the necessary information for authenticating with OpenSearch.
type Authenticator struct {
	clientTransport opensearchapi.Transport
	config          config.OpensearchClientConfig
	tokenReceiver   ITokenReceiver
	authMethod      authMethod
}

// Ensure Authenticator implements the opensearchapi.Transport interface.
var _ opensearchapi.Transport = &Authenticator{}

// InjectAuthenticationIntoClient is a function that sets up the authentication method for the OpenSearch client.
// client is the OpenSearch client to inject the authentication into.
// config is the configuration for the OpenSearch client.
// tokenReceiver is the token receiver for OpenID authentication and must implement the GetClientAccessToken function. It can be nil for basic authentication.
func InjectAuthenticationIntoClient(client *opensearch.Client, config config.OpensearchClientConfig, tokenReceiver ITokenReceiver) error {
	method, err := getAuthenticationMethod(config, tokenReceiver)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Set up auth method for opensearch client: %s", method)

	authenticator := &Authenticator{
		config:        config,
		tokenReceiver: tokenReceiver,
		authMethod:    method,
	}

	// store the original Transport interface implementation of the opensearch client to be able to wrap it
	// in the authenticator.Perform method
	authenticator.clientTransport = client.Transport
	// replace the original Transport interface implementation of the opensearch client with the wrapper
	client.Transport = authenticator

	return nil
}

// getAuthenticationMethod is a helper function that determines the authentication method based on the provided configuration.
func getAuthenticationMethod(conf config.OpensearchClientConfig, tokenReceiver ITokenReceiver) (authMethod, error) {
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

// Perform is a method that implements the opensearchapi.Transport interface.
// It injects the authentication header into the request and then performs the request.
func (a *Authenticator) Perform(req *http.Request) (*http.Response, error) {
	requestWithInjectedAuth, err := a.injectAuthenticationHeader(req)
	if err != nil {
		return nil, err
	}
	return a.clientTransport.Perform(requestWithInjectedAuth)
}
