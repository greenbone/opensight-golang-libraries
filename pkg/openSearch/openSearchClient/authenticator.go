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
	basic authMethod = "basic auth"
	openId = "openID"
	none = "none"
)

type Authenticator struct {
	Transport     opensearchapi.Transport
	config        config.OpensearchClientConfig
	tokenReceiver ITokenReceiver
	authMethod    authMethod
}

var _ esapi.Transport = &Authenticator{}

func InjectAuthenticationIntoClient(client *opensearch.Client, config config.OpensearchClientConfig, tokenReceiver ITokenReceiver) {
	c := &Authenticator{
		config:        config,
		tokenReceiver: tokenReceiver,
		authMethod:    determineAuthenticationMethod(config, tokenReceiver),
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
	return config.KeycloakClient != "" && config.KeycloakClientSecret != ""
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
	switch method {
	case basic:
		log.Debug().Msgf("opensearch basic auth")
		reqClone.SetBasicAuth(c.config.Username, c.config.Password)
	case openId:
		log.Debug().Msgf("opensearch openID auth")
		token, err := c.tokenReceiver.GetClientAccessToken(c.config.KeycloakClient, c.config.KeycloakClientSecret)
		if err != nil {
			log.Error().Msgf("Could not retrieve authorization header: %v", err)
			return reqClone
		}
		reqClone.Header.Set("Authorization", "Bearer "+token)
	case none:
		fallthrough
	default:
		log.Debug().Msgf("opensearch no auth")
	}
	return reqClone
}

// Perform implements the opensearchapi.Transport interface
func (c *Authenticator) Perform(req *http.Request) (*http.Response, error) {
	requestWithInjectedAuth := c.injectAuthenticationHeader(c.authMethod, req)
	return c.Transport.Perform(requestWithInjectedAuth)
}
