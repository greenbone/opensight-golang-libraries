// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package httpassert

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter(t *testing.T) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprint(w, `{"asd":"123","items":["a","b","c"]}`)
		require.NoError(t, err)
	})

	mux.HandleFunc("/json/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprint(w, `{"asd":"123","items":["a","b","c"]}`)
		require.NoError(t, err)
	})

	mux.HandleFunc("/text", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, err := fmt.Fprint(w, "ok")
		require.NoError(t, err)
	})

	mux.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {})

	return mux
}

func TestRequestPerform(t *testing.T) {
	router := setupRouter(t)
	request := New(t, router)

	tests := map[string]struct {
		verb    string
		content string
	}{
		"Get": {
			verb: "Get",
		},
		"Post": {
			verb:    "Post",
			content: `{}`,
		},
		"Put": {
			verb:    "Put",
			content: `{}`,
		},
		"Delete": {
			verb: "Delete",
		},
		"Options": {
			verb: "Options",
		},
		"Get (all caps)": {
			verb: http.MethodGet,
		},
		"Post (all caps)": {
			verb:    http.MethodPost,
			content: `{}`,
		},
		"Put (all caps)": {
			verb:    http.MethodPut,
			content: `{}`,
		},
		"Delete (all caps)": {
			verb: http.MethodDelete,
		},
		"Options (all caps)": {
			verb: http.MethodOptions,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			request.Perform(tc.verb, "/json").
				Content(tc.content).
				Expect().
				StatusCode(http.StatusOK)
		})

		t.Run(name+"f", func(t *testing.T) {
			request.Performf(tc.verb, "/json/%s", "foo").
				Content(tc.content).
				Expect().
				StatusCode(http.StatusOK)
		})
	}
}

func TestRequest(t *testing.T) {
	router := setupRouter(t)
	request := New(t, router)

	t.Run("AuthHeader sets header", func(t *testing.T) {
		var gotAuth string
		var gotFoo string
		router := http.NewServeMux()
		router.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			gotFoo = r.Header.Get("foo")
			w.WriteHeader(http.StatusOK)
		})

		New(t, router).Get("/auth").
			Headers(map[string]string{
				"foo": "bar",
			}).
			AuthHeader("Bearer abc").
			Expect().
			StatusCode(http.StatusOK)

		if gotAuth != "Bearer abc" {
			t.Errorf("expected Authorization header 'Bearer abc', got %q", gotAuth)
		}
		if gotFoo != "bar" {
			t.Errorf("expected Authorization header 'foo', got %q", gotFoo)
		}
	})

	t.Run("Auth JWT", func(t *testing.T) {
		var gotAuth string
		router := http.NewServeMux()
		router.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		})

		New(t, router).Get("/auth").
			AuthJwt("test JWT").
			Expect().
			StatusCode(http.StatusOK)

		assert.Equal(t, `bearer test JWT`, gotAuth)
	})

	t.Run("ContentFile loads file content", func(t *testing.T) {
		path := writeTempFile(t, `{"foo":"bar"}`)
		request.Post("/json").
			ContentFile(path).
			ContentType("application/json").
			Expect().
			StatusCode(http.StatusOK)
	})

	t.Run("JSON content template", func(t *testing.T) {
		var content []byte
		router := http.NewServeMux()
		router.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			content = bodyBytes
			w.WriteHeader(http.StatusOK)
		})

		New(t, router).
			Post("/json").
			JsonContentTemplate(`{
				"n": {
					"foo": "bar",
					"asd": ""
				}
			}`, map[string]any{
				"$.n.asd": "123",
			}).
			Expect().
			StatusCode(http.StatusOK)

		assert.JSONEq(t, `{"n": {"foo": "bar","asd":"123"}}`, string(content))
	})
}

func TestExpectEventually(t *testing.T) {
	t.Parallel()

	counter := 0
	router := http.NewServeMux()
	router.HandleFunc("/api/jobs/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if counter < 3 {
			t.Logf("counter: %d", counter)
			counter++
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprint(w, `{"data":{"status":"running"}}`)
			require.NoError(t, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, `{"data":{"status":"finished"}}`)
		require.NoError(t, err)
	})

	request := New(t, router)

	request.Get("/api/jobs/123").
		ExpectEventually(func(r Response) {
			r.StatusCode(http.StatusOK).
				JsonPath("$.data.status", "finished")
		}, 1*time.Second, 20*time.Millisecond)
}
