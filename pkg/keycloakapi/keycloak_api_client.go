// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package keycloakapi

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
	Name string `json:"name"` // Name represents (always) the full group path
}

// User represents a Keycloak user.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// Internal DTO for Keycloak response.
type keycloakGroup struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Path      string          `json:"path"`
	SubGroups []keycloakGroup `json:"subGroups"`
}

// KeycloakAPIClient provides methods to interact with Keycloak's REST API.
type KeycloakAPIClient struct {
	authClient *auth.KeycloakClient
}

// NewKeycloakAPIClient creates a new KeycloakAPIClient.
func NewKeycloakAPIClient(authClient *auth.KeycloakClient) *KeycloakAPIClient {
	return &KeycloakAPIClient{authClient: authClient}
}

// ListGroups retrieves all groups from Keycloak.
// ListGroups retrieves all groups from Keycloak and flattens groups + subgroups.
func (kc *KeycloakAPIClient) ListGroups(ctx context.Context) ([]Group, error) {
	token, err := kc.authClient.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	config := kc.authClient.Config()
	url := fmt.Sprintf("%s/admin/realms/%s/groups?search=*", config.AuthURL, config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := kc.authClient.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to list groups: %s: could not read response body: %w", resp.Status, err)
		}
		return nil, fmt.Errorf("failed to list groups: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var rootGroups []keycloakGroup
	if err := json.NewDecoder(resp.Body).Decode(&rootGroups); err != nil {
		return nil, fmt.Errorf("failed to decode groups response: %w", err)
	}

	flattened := make([]Group, 0)
	for _, g := range rootGroups {
		flattened = append(flattened, flattenGroupTree(g)...)
	}

	return flattened, nil
}

// ListUsers retrieves all users from Keycloak.
func (kc *KeycloakAPIClient) ListUsers(ctx context.Context) ([]User, error) {
	token, err := kc.authClient.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	config := kc.authClient.Config()
	url := fmt.Sprintf("%s/admin/realms/%s/users", config.AuthURL, config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := kc.authClient.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %s: could not read response body: %w", resp.Status, err)
		}

		return nil, fmt.Errorf("failed to list users: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode users response: %w", err)
	}
	return users, nil
}

func flattenGroupTree(g keycloakGroup) []Group {
	name := g.Path // Name represents (always) the full group path,

	out := []Group{
		{
			ID:   g.ID,
			Name: name, // Use path as requested
		},
	}

	for _, child := range g.SubGroups {
		out = append(out, flattenGroupTree(child)...)
	}

	return out
}
