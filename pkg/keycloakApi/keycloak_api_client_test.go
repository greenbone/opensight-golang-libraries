// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package keycloakApi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/auth"
	"github.com/stretchr/testify/require"
)

func newTestKeycloakClient(serverURL string) *auth.KeycloakClient {
	return auth.NewKeycloakClient(
		&http.Client{},
		auth.KeycloakConfig{
			AuthURL: serverURL,
			Realm:   "test-realm",
		},
		auth.ClientCredentials{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		},
	)
}

func TestListGroups(t *testing.T) {
	groupsJSON := `[{"id":"1","name":"group1"},{"id":"2","name":"group2"}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/realms/test-realm/protocol/openid-connect/token" {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"access_token":"test-token","expires_in":3600}`))
			require.NoError(t, err)
			return
		}
		if r.URL.Path == "/admin/realms/test-realm/groups" {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(groupsJSON))
			require.NoError(t, err)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	authClient := newTestKeycloakClient(server.URL)
	apiClient := NewKeycloakAPIClient(authClient)
	groups, err := apiClient.ListGroups(context.Background())
	require.NoError(t, err, "expected no error")

	require.Equal(t, 2, len(groups), "expected 2 groups")
	require.Equal(t, "group1", groups[0].Name, "first group name mismatch")
	require.Equal(t, "group2", groups[1].Name, "second group name mismatch")
}

func TestListUsers(t *testing.T) {
	usersJSON := `[{"id":"1","username":"user1","email":"user1@example.com"},{"id":"2","username":"user2","email":"user2@example.com"}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/realms/test-realm/protocol/openid-connect/token" {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"access_token":"test-token","expires_in":3600}`))
			require.NoError(t, err)
			return
		}
		if r.URL.Path == "/admin/realms/test-realm/users" {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(usersJSON))
			require.NoError(t, err)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	authClient := newTestKeycloakClient(server.URL)
	apiClient := NewKeycloakAPIClient(authClient)
	users, err := apiClient.ListUsers(context.Background())
	require.NoError(t, err, "expected no error")

	require.Equal(t, 2, len(users), "expected 2 users")
	require.Equal(t, "user1", users[0].Username, "first user username mismatch")
	require.Equal(t, "user2@example.com", users[1].Email, "second user email mismatch")
}
