package openSearchClient

import (
	"context"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/rs/zerolog/log"
	"net/http"
)

type AuthClient struct {
	*opensearch.Client
	config        *config.OpensearchClientConfig
	tokenReceiver ITokenReceiver
	authMethod    authMethod
}

var _ esapi.Transport = &AuthClient{}

func NewAuthClient(client *opensearch.Client, config *config.OpensearchClientConfig, tokenReceiver ITokenReceiver) *AuthClient {
	c := &AuthClient{
		Client:        client,
		config:        config,
		tokenReceiver: tokenReceiver,
		authMethod:    determineAuthenticationMethod(config, tokenReceiver),
	}

	c.InjectAuthenticationIntoClient()
	return c
}

func (c *AuthClient) InjectAuthenticationIntoClient() {
	c.Client.Transport = c
}

func isEligableForOpenId(config *config.OpensearchClientConfig) bool {
	return config.KeycloakClient != "" && config.KeycloakClientSecret != ""
}

func determineAuthenticationMethod(config *config.OpensearchClientConfig, tokenReceiver ITokenReceiver) authMethod {
	if config.Username != "" && config.Password != "" {
		return basic
	} else if tokenReceiver != nil && isEligableForOpenId(config) {
		return openId
	} else {
		return none
	}
}

// Http Transport middleware

func (c *AuthClient) injectAuthenticationHeader(method authMethod, req *http.Request) *http.Request {
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

func (c *AuthClient) Perform(req *http.Request) (*http.Response, error) {
	log.Debug().Msgf("Auth method %v", c.authMethod)
	requestWithInjectedAuth := c.injectAuthenticationHeader(c.authMethod, req)
	return c.Client.Perform(requestWithInjectedAuth)
}
