package openSearchClient

import (
	"context"
	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/rs/zerolog/log"
	"net/http"
)

// TODO refactor, make generic

// Base - For all

func isEligableForOpenId(config config.OpensearchClientConfig) bool {
	return config.KeycloakClient != "" && config.KeycloakClientSecret != ""
}

func (c *Client) determineAuthenticationMethod() authMethod {
	if c.config.Username != "" && c.config.Password != "" {
		return basic
	} else if c.tokenReceiver != nil && isEligableForOpenId(c.config) {
		return openId
	} else {
		return none
	}
}

// Http Transport middleware

func (c *Client) injectAuthenticationHeader(method authMethod, req *http.Request) *http.Request {
	reqClone := req.Clone(context.Background())
	switch method {
	case basic:
		log.Debug().Msgf("opensearch basic auth")
		reqClone.SetBasicAuth(c.config.Username, c.config.Password)
	case openId:
		log.Debug().Msgf("opensearch openID auth")
		token, err := c.tokenReceiver.GetClientAccessToken(c.config.KeycloakClient, c.config.KeycloakClientSecret)
		if err != nil {
			log.Error().Err(err).Msgf("Could not retrieve authorization header: %s", err)
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

func (c *Client) Perform(req *http.Request) (*http.Response, error) {
	method := c.determineAuthenticationMethod()
	log.Debug().Msgf("Auth method %v", method)
	requestWithInjectedAuth := c.injectAuthenticationHeader(method, req)
	return c.openSearchProjectClient.Perform(requestWithInjectedAuth)
}

// OpenSearch Search middleware
func (c *Client) injectAuthenticationHeaderSearchRequest(method authMethod, req *opensearchapi.SearchRequest) *opensearchapi.SearchRequest {
	switch method {
	case basic:
		log.Debug().Msgf("opensearch basic auth")
	case openId:
		log.Debug().Msgf("opensearch openID auth")
		token, err := c.tokenReceiver.GetClientAccessToken(c.config.KeycloakClient, c.config.KeycloakClientSecret)
		if err != nil {
			log.Error().Err(err).Msgf("Could not retrieve authorization header: %s", err)
			return req
		}
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header.Add("Authorization", "Bearer "+token)

	case none:
		fallthrough
	default:
		log.Debug().Msgf("opensearch no auth")
	}
	return req
}

func (c *Client) SearchAuthenticationMiddleware(req *opensearchapi.SearchRequest) {
	method := c.determineAuthenticationMethod()
	log.Debug().Msgf("Auth method %v", method)
	c.injectAuthenticationHeaderSearchRequest(method, req)
}

// OpenSearch deleteByQuery Middleware

func (c *Client) injectAuthenticationHeaderDeleteByQueryRequest(method authMethod, req *opensearchapi.DeleteByQueryRequest) *opensearchapi.DeleteByQueryRequest {
	switch method {
	case basic:
		log.Debug().Msgf("opensearch basic auth")
	case openId:
		log.Debug().Msgf("opensearch openID auth")
		token, err := c.tokenReceiver.GetClientAccessToken(c.config.KeycloakClient, c.config.KeycloakClientSecret)
		if err != nil {
			log.Error().Err(err).Msgf("Could not retrieve authorization header: %s", err)
			return req
		}
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header.Add("Authorization", "Bearer "+token)

	case none:
		fallthrough
	default:
		log.Debug().Msgf("opensearch no auth")
	}
	return req
}

func (c *Client) DeleteByQueryAuthenticationMiddleware(req *opensearchapi.DeleteByQueryRequest) {
	method := c.determineAuthenticationMethod()
	log.Debug().Msgf("Auth method %v", method)
	c.injectAuthenticationHeaderDeleteByQueryRequest(method, req)
}

// OpenSearch Bulk Middleware

func (c *Client) injectAuthenticationHeaderBulkRequest(method authMethod, req *opensearchapi.BulkRequest) *opensearchapi.BulkRequest {
	switch method {
	case basic:
		log.Debug().Msgf("opensearch basic auth")
	case openId:
		log.Debug().Msgf("opensearch openID auth")
		token, err := c.tokenReceiver.GetClientAccessToken(c.config.KeycloakClient, c.config.KeycloakClientSecret)
		if err != nil {
			log.Error().Err(err).Msgf("Could not retrieve authorization header: %s", err)
			return req
		}
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header.Add("Authorization", "Bearer "+token)

	case none:
		fallthrough
	default:
		log.Debug().Msgf("opensearch no auth")
	}
	return req
}

func (c *Client) BulkAuthenticationMiddleware(req *opensearchapi.BulkRequest) {
	method := c.determineAuthenticationMethod()
	log.Debug().Msgf("Auth method %v", method)
	c.injectAuthenticationHeaderBulkRequest(method, req)
}
