// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/rs/zerolog/log"
)

type authMethod string

const (
	basic  authMethod = "basic"
	openId authMethod = "openid"
)

type ITokenReceiver interface {
	GetClientAccessToken(clientName, clientSecret string) (string, error)
}

type Authenticator struct {
	clientTransport opensearchapi.Transport
	config          config.OpensearchClientConfig
	tokenReceiver   ITokenReceiver
	authMethod      authMethod
}

var _ esapi.Transport = &Authenticator{}

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

func validateOpenId(tokenReceiver ITokenReceiver) error {
	if tokenReceiver == nil {
		return fmt.Errorf("token receiver must not be nil for openid authentication")
	}
	return nil
}

func getAuthenticationMethod(conf config.OpensearchClientConfig, tokenReceiver ITokenReceiver) (authMethod, error) {
	if conf.AuthUsername == "" || conf.AuthPassword == "" {
		return "", fmt.Errorf("username and password must be set in configuration")
	}
	method := authMethod(strings.ToLower(conf.AuthMethod))

	switch method {
	case basic:
	case openId:
		if err := validateOpenId(tokenReceiver); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("invalid authentication method for opensearch: %s", conf.AuthMethod)
	}

	return method, nil
}

func (a *Authenticator) injectAuthenticationHeader(method authMethod, req *http.Request) *http.Request {
	reqClone := req.Clone(context.Background())
	if reqClone.Header == nil {
		reqClone.Header = http.Header{}
	}
	switch method {
	case basic:
		reqClone.SetBasicAuth(a.config.AuthUsername, a.config.AuthPassword)
	case openId:
		token, err := a.tokenReceiver.GetClientAccessToken(a.config.AuthUsername, a.config.AuthPassword)
		if err != nil {
			log.Error().Msgf("Could not retrieve authorization header: %v", err)
			return reqClone
		}
		reqClone.Header.Set("Authorization", "Bearer "+token)
	default:
		log.Error().Msgf("undefined authentication method for opensearch client: %s", method)
	}
	return reqClone
}

// Perform implements the opensearchapi.Transport interface
func (a *Authenticator) Perform(req *http.Request) (*http.Response, error) {
	requestWithInjectedAuth := a.injectAuthenticationHeader(a.authMethod, req)
	return a.clientTransport.Perform(requestWithInjectedAuth)
}
