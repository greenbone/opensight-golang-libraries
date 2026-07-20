// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package keycloakapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/auth"
	"github.com/stretchr/testify/assert"
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

	wantGroups := []Group{
		{ID: "1", Name: "group1"},
		{ID: "2", Name: "group2"},
	}
	assert.ElementsMatch(t, wantGroups, groups, "groups mismatch")
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

	wantUsers := []User{
		{ID: "1", Username: "user1", Email: "user1@example.com"},
		{ID: "2", Username: "user2", Email: "user2@example.com"},
	}
	assert.ElementsMatch(t, wantUsers, users, "users mismatch")
}

func TestListGroups_FlattensNestedGroups_UsesPathAsName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/realms/test-realm/protocol/openid-connect/token":
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"access_token":"test-token","expires_in":3600}`))
			require.NoError(t, err)
			return

		case "/admin/realms/test-realm/groups":
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`[
				  {
					"id":"p1",
					"name":"Platform",
					"path":"/Platform",
					"subGroupCount":2,
					"subGroups":[
					  {"id":"c1","name":"Admins","path":"/Platform/Admins","subGroupCount":0,"subGroups":[]},
					  {"id":"c2","name":"Users","path":"/Platform/Users","subGroupCount":0,"subGroups":[
					  	{"id":"g1","name":"Grandkid","path":"/Platform/Users/Grandkid","subGroupCount":0,"subGroups":[]}]}
					]
				  },
				  {
					"id":"s1",
					"name":"System",
					"path":"/System",
					"subGroupCount":0,
					"subGroups":[]
				  }
				]`))
			require.NoError(t, err)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	authClient := newTestKeycloakClient(server.URL)
	apiClient := NewKeycloakAPIClient(authClient)

	got, err := apiClient.ListGroups(context.Background())
	require.NoError(t, err)

	want := []Group{
		{ID: "p1", Name: "/Platform"},
		{ID: "c1", Name: "/Platform/Admins"},
		{ID: "c2", Name: "/Platform/Users"},
		{ID: "g1", Name: "/Platform/Users/Grandkid"},
		{ID: "s1", Name: "/System"},
	}
	assert.ElementsMatch(t, want, got)
}
