// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package auth provides a client to authenticate against a Keycloak server.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

const tokenRefreshMargin = 10 * time.Second

// KeycloakConfig holds the credentials and configuration details
type KeycloakConfig struct {
	AuthURL string
	Realm   string
}

type tokenInfo struct {
	AccessToken string //nolint:gosec
	ExpiresAt   time.Time
}

type authResponse struct {
	AccessToken string `json:"access_token"` //nolint:gosec
	ExpiresIn   int    `json:"expires_in"`   // in seconds
}

// Credentials holds the required credentials and determines the used auth type.
type Credentials interface {
	constructUrlValues() url.Values // unexported method to prevent external implementation
}

// ClientCredentials to authenticate via `Client credentials grant` flow.
// Ref: https://www.keycloak.org/docs/latest/server_admin/index.html#_client_credentials_grant
type ClientCredentials struct {
	ClientID     string
	ClientSecret string
}

func (c ClientCredentials) constructUrlValues() url.Values {
	urlValues := url.Values{}
	urlValues.Set("grant_type", "client_credentials")
	urlValues.Set("client_id", c.ClientID)
	urlValues.Set("client_secret", c.ClientSecret)

	return urlValues
}

// ResourceOwnerCredentials to authenticate via `Resource owner password credentials grant` flow.
// Ref: https://www.keycloak.org/docs/latest/server_admin/index.html#_oidc-auth-flows-direct
type ResourceOwnerCredentials struct {
	ClientID string
	Username string
	Password string
}

func (c ResourceOwnerCredentials) constructUrlValues() url.Values {
	urlValues := url.Values{}
	urlValues.Set("grant_type", "password")
	urlValues.Set("client_id", c.ClientID)
	urlValues.Set("username", c.Username)
	urlValues.Set("password", c.Password)

	return urlValues
}

// KeycloakClient can be used to authenticate against a Keycloak server.
type KeycloakClient struct {
	httpClient  *http.Client
	cfg         KeycloakConfig
	credentials Credentials
	tokenInfo   tokenInfo
	tokenMutex  sync.RWMutex

	clock Clock // to mock time in tests
}

// NewKeycloakClient creates a new KeycloakClient. Passed [Credentials] determines the used auth type.
func NewKeycloakClient(httpClient *http.Client, cfg KeycloakConfig, credentials Credentials) *KeycloakClient {
	return &KeycloakClient{
		httpClient:  httpClient,
		cfg:         cfg,
		credentials: credentials,
		clock:       realClock{},
	}
}

// GetToken retrieves a valid access token. The token is cached and refreshed before expiry.
func (c *KeycloakClient) GetToken(ctx context.Context) (string, error) {
	token, ok := c.getCachedToken()
	if ok {
		return token, nil
	}

	// need to retrieve new token
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// check again in case another goroutine already refreshed the token
	if c.clock.Now().Before(c.tokenInfo.ExpiresAt.Add(-tokenRefreshMargin)) {
		return c.tokenInfo.AccessToken, nil
	}

	authResponse, err := c.requestToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}

	c.tokenInfo = tokenInfo{
		AccessToken: authResponse.AccessToken,
		ExpiresAt:   c.clock.Now().UTC().Add(time.Duration(authResponse.ExpiresIn) * time.Second),
	}

	return authResponse.AccessToken, nil
}

func (c *KeycloakClient) getCachedToken() (token string, ok bool) {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	if c.clock.Now().Before(c.tokenInfo.ExpiresAt.Add(-tokenRefreshMargin)) {
		return c.tokenInfo.AccessToken, true
	}
	return "", false
}

func (c *KeycloakClient) requestToken(ctx context.Context) (*authResponse, error) {
	data := c.credentials.constructUrlValues()

	authenticationURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		c.cfg.AuthURL, c.cfg.Realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authenticationURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to execute authentication request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("authentication request failed with status: %s: %w body: %s", resp.Status, err, string(respBody))
		}
		return nil, fmt.Errorf("authentication request failed with status: %s: %s", resp.Status, string(respBody))
	}

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to parse authentication response: %w", err)
	}

	return &authResp, nil
}

// HTTPClient returns the underlying http.Client used by the KeycloakClient.
func (c *KeycloakClient) HTTPClient() *http.Client {
	return c.httpClient
}

// Config returns the KeycloakConfig used by the KeycloakClient.
func (c *KeycloakClient) Config() KeycloakConfig {
	return c.cfg
}
