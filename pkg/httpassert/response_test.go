// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package httpassert

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse(t *testing.T) {
	router := setupRouter(t)
	request := New(t, router)

	t.Run("JsonFile assertion", func(t *testing.T) {
		path := writeTempFile(t, `{"asd":"123","items":["a","b","c"]}`)
		request.Post("/json").
			Expect().
			JsonFile(path)
	})

	t.Run("JsonTemplate value replacement and ignore handling", func(t *testing.T) {
		router := http.NewServeMux()
		router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprint(w, `{"id":42,"name":"Greenbone","version":"1.0.0","tags":["alpha","beta"]}`)
			require.NoError(t, err)
		})

		m := New(t, router)

		t.Run("replaces scalar and nested array values", func(t *testing.T) {
			m.Get("/api").
				Expect().
				JsonTemplate(`{"id":0,"name":"","version":"","tags":["x","y"]}`, map[string]any{
					"$.id":      42,
					"$.name":    "Greenbone",
					"$.tags[0]": "alpha",
					"$.tags[1]": "beta",
					"$.version": "1.0.0",
				})
		})

		t.Run("ignores values marked as <IGNORE>", func(t *testing.T) {
			m.Get("/api").
				Expect().
				JsonTemplate(`{"id":0,"name":"","version":"","tags":["x","y"]}`, map[string]any{
					"$.id":      42,
					"$.name":    "Greenbone",
					"$.version": IgnoreJsonValue, // should not cause mismatch
					"$.tags[0]": "alpha",
					"$.tags[1]": "beta",
				})
		})

		t.Run("handles nested structure with mixed types", func(t *testing.T) {
			router := http.NewServeMux()
			router.HandleFunc("/complex", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{"meta":{"id":10,"info":{"env":"prod","build":17}}}`))
				require.NoError(t, err)
			})

			New(t, router).Get("/complex").
				Expect().
				JsonTemplate(`{"meta":{"id":0,"info":{"env":"","build":0}}}`, map[string]any{
					"$.meta.id":         10,
					"$.meta.info.env":   "prod",
					"$.meta.info.build": IgnoreJsonValue, // ignored build number
				})
		})
	})

	t.Run("GetBody returns non-empty", func(t *testing.T) {
		body := request.Post("/json").
			Expect().
			GetBody()
		if body == "" {
			t.Error("expected non-empty body")
		}
	})

	t.Run("GetJsonBodyObject unmarshal data", func(t *testing.T) {
		type foo struct {
			Asd   string   `json:"asd"`
			Items []string `json:"items"`
		}
		asd := foo{}

		request.Post("/json").
			Expect().
			GetJsonBodyObject(&asd)

		assert.Equal(t, "123", asd.Asd)
		assert.Equal(t, []string{"a", "b", "c"}, asd.Items)
	})

	t.Run("Log prints response", func(t *testing.T) {
		request.Post("/json").
			Expect().
			Log()
	})

	t.Run("JsonTemplateFile", func(t *testing.T) {
		router := http.NewServeMux()
		router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprint(w, `{"id":42,"name":"Greenbone","version":"1.0.0"}`)
			require.NoError(t, err)
		})

		m := New(t, router)

		t.Run("successfully compares JSON template file", func(t *testing.T) {
			path := writeTempFile(t, `{"id":0,"name":"","version":""}`)

			m.Get("/api").
				Expect().
				JsonTemplateFile(path, map[string]any{
					"$.id":      42,
					"$.name":    "Greenbone",
					"$.version": "1.0.0",
				})
		})

		t.Run("handles <IGNORE> value in JsonTemplateFile", func(t *testing.T) {
			path := writeTempFile(t, `{"id":0,"name":"","version":""}`)

			m.Get("/api").
				Expect().
				JsonTemplateFile(path, map[string]any{
					"$.id":      42,
					"$.name":    "Greenbone",
					"$.version": IgnoreJsonValue, // ignored field
				})
		})
	})

	t.Run("Header", func(t *testing.T) {
		router := http.NewServeMux()
		router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})

		t.Run("compare string", func(t *testing.T) {
			m := New(t, router)

			m.Get("/api").
				Expect().
				Header("Content-Type", "application/json")
		})

		t.Run("extract value to variable", func(t *testing.T) {
			m := New(t, router)

			var value string
			m.Get("/api").
				Expect().
				Header("Content-Type", ExtractTo(&value))
			assert.Equal(t, "application/json", value)
		})

		t.Run("use matcher", func(t *testing.T) {
			m := New(t, router)

			m.Get("/api").
				Expect().
				Header("Content-Type", Regex("application/[json]"))
		})
	})

	t.Run("POST basic JSON", func(t *testing.T) {
		request.Post("/json").
			ContentType("application/json").
			Content(`{"asd":"123"}`).
			Expect().
			StatusCode(http.StatusOK).
			ContentType("application/json").
			JsonPath("$.asd", "123").
			Json(`{"asd":"123","items":["a","b","c"]}`)
	})

	t.Run("Get with list JSON", func(t *testing.T) {
		request.Get("/json").
			Expect().
			StatusCode(http.StatusOK).
			JsonPath("$.items[0]", "a").
			JsonPathJson("$.items", `["a","b","c"]`)
	})

	t.Run("GET empty body", func(t *testing.T) {
		request.Get("/empty").
			Expect().
			StatusCode(http.StatusOK).
			NoContent()
	})

	t.Run("plain text", func(t *testing.T) {
		request.Put("/text").
			ContentType("text/plain").
			Content("ok").
			Expect().
			StatusCode(http.StatusOK).
			ContentType("text/plain").
			Body("ok")
	})
}
