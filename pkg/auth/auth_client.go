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

const tokenRefreshMargin = 10 * time.Second

// KeycloakConfig holds the credentials and configuration details
type KeycloakConfig struct {
	ClientID      string
	Username      string
	Password      string
	AuthURL       string
	KeycloakRealm string
}

type tokenInfo struct {
	AccessToken string
	ExpiresAt   time.Time
}

type authResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // in seconds
}

// KeycloakClient can be used to authenticate against a Keycloak server.
type KeycloakClient struct {
	httpClient *http.Client
	cfg        KeycloakConfig
	tokenInfo  tokenInfo
	tokenMutex sync.RWMutex
}

// NewKeycloakClient creates a new KeycloakClient.
func NewKeycloakClient(httpClient *http.Client, cfg KeycloakConfig) *KeycloakClient {
	return &KeycloakClient{
		httpClient: httpClient,
		cfg:        cfg,
	}
}

// GetToken retrieves a valid access token. The token is cached and refreshed before expiry.
// The token is obtained by `Resource owner password credentials grant` flow.
// Ref: https://www.keycloak.org/docs/latest/server_admin/index.html#_oidc-auth-flows-direct
func (c *KeycloakClient) GetToken(ctx context.Context) (string, error) {
	getCachedToken := func() (token string, ok bool) {
		c.tokenMutex.RLock()
		defer c.tokenMutex.RUnlock()
		if time.Now().Before(c.tokenInfo.ExpiresAt.Add(-tokenRefreshMargin)) {
			return c.tokenInfo.AccessToken, true
		} else {
			return "", false
		}
	}

	token, ok := getCachedToken()
	if ok {
		return token, nil
	}

	// need to retrieve new token
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// check again in case another goroutine already refreshed the token
	if time.Now().Before(c.tokenInfo.ExpiresAt.Add(-tokenRefreshMargin)) {
		return c.tokenInfo.AccessToken, nil
	}

	data := url.Values{}
	data.Set("client_id", c.cfg.ClientID)
	data.Set("password", c.cfg.Password)
	data.Set("username", c.cfg.Username)
	data.Set("grant_type", "password")

	authenticationURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		c.cfg.AuthURL, c.cfg.KeycloakRealm)

	req, err := http.NewRequest(http.MethodPost, authenticationURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create authentication request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute authentication request with retry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			respBody = []byte("failed to read response body: " + err.Error())
		}
		return "", fmt.Errorf("authentication request failed with status: %s: %s", resp.Status, string(respBody))
	}

	var authResponse authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("failed to parse authentication response: %w", err)
	}

	c.tokenInfo = tokenInfo{
		AccessToken: authResponse.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(authResponse.ExpiresIn) * time.Second),
	}

	return authResponse.AccessToken, nil
}
