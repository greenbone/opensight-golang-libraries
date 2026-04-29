// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package keycloakApi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/auth"
)

// Group represents a Keycloak group.
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// User represents a Keycloak user.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// KeycloakAPIClient provides methods to interact with Keycloak's REST API.
type KeycloakAPIClient struct {
	AuthClient *auth.KeycloakClient
}

// NewKeycloakAPIClient creates a new KeycloakAPIClient.
func NewKeycloakAPIClient(authClient *auth.KeycloakClient) *KeycloakAPIClient {
	return &KeycloakAPIClient{AuthClient: authClient}
}

// ListGroups retrieves all groups from Keycloak.
func (kc *KeycloakAPIClient) ListGroups(ctx context.Context) ([]Group, error) {
	token, err := kc.AuthClient.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	config := kc.AuthClient.Config()
	url := fmt.Sprintf("%s/admin/realms/%s/groups", config.AuthURL, config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := kc.AuthClient.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list groups: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var groups []Group
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("failed to decode groups response: %w", err)
	}
	return groups, nil
}

// ListUsers retrieves all users from Keycloak.
func (kc *KeycloakAPIClient) ListUsers(ctx context.Context) ([]User, error) {
	token, err := kc.AuthClient.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	config := kc.AuthClient.Config()
	url := fmt.Sprintf("%s/admin/realms/%s/users", config.AuthURL, config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := kc.AuthClient.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list users: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode users response: %w", err)
	}
	return users, nil
}
