// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"context"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/rs/zerolog/log"
)

type authMethod string

const (
	basic  authMethod = "basic auth"
	openId authMethod = "openID"
	none   authMethod = "none"
)

type ITokenReceiver interface {
	GetClientAccessToken(clientName, clientSecret string) (string, error)
}

type Authenticator struct {
	Transport     opensearchapi.Transport
	config        config.OpensearchClientConfig
	tokenReceiver ITokenReceiver
	authMethod    authMethod
}

var _ esapi.Transport = &Authenticator{}

func InjectAuthenticationIntoClient(client *opensearch.Client, config config.OpensearchClientConfig, tokenReceiver ITokenReceiver) {
	method := determineAuthenticationMethod(config, tokenReceiver)
	log.Debug().Msgf("Set up auth method for opensearch client: %s", method)

	c := &Authenticator{
		config:        config,
		tokenReceiver: tokenReceiver,
		authMethod:    method,
	}

	// store the original Transport interface implementation of the opensearch client to be able to wrap it
	c.Transport = client.Transport
	// replace the original Transport interface implementation with the wrapper
	client.Transport = c
}

func isEligibleForBasicAuth(config config.OpensearchClientConfig) bool {
	return config.Username != "" && config.Password != ""
}

func isEligibleForOpenId(config config.OpensearchClientConfig) bool {
	return config.IDPClientID != "" && config.IDPClientSecret != ""
}

func determineAuthenticationMethod(config config.OpensearchClientConfig, tokenReceiver ITokenReceiver) authMethod {
	if isEligibleForBasicAuth(config) {
		return basic
	} else if tokenReceiver != nil && isEligibleForOpenId(config) {
		return openId
	} else {
		return none
	}
}

func (c *Authenticator) injectAuthenticationHeader(method authMethod, req *http.Request) *http.Request {
	reqClone := req.Clone(context.Background())
	if reqClone.Header == nil {
		reqClone.Header = http.Header{}
	}
	switch method {
	case basic:
		reqClone.SetBasicAuth(c.config.Username, c.config.Password)
	case openId:
		token, err := c.tokenReceiver.GetClientAccessToken(c.config.IDPClientID, c.config.IDPClientSecret)
		if err != nil {
			log.Error().Msgf("Could not retrieve authorization header: %v", err)
			return reqClone
		}
		reqClone.Header.Set("Authorization", "Bearer "+token)
	case none:
	default:
		log.Error().Msgf("undefined authentication method for opensearch client: %s", method)
	}
	return reqClone
}

// Perform implements the opensearchapi.Transport interface
func (c *Authenticator) Perform(req *http.Request) (*http.Response, error) {
	requestWithInjectedAuth := c.injectAuthenticationHeader(c.authMethod, req)
	return c.Transport.Perform(requestWithInjectedAuth)
}
