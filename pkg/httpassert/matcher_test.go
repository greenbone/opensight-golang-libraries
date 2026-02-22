// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package httpassert

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_HasSizeMatcher(t *testing.T) {
	router := http.NewServeMux()
	router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, `{"data":[1,2]}`)
		assert.NoError(t, err)
	})

	request := New(t, router)

	request.Get("/api").
		Expect().
		StatusCode(http.StatusOK).
		JsonPath("$.data", HasSize(2))
}

func Test_ContainsMatcher(t *testing.T) {
	router := http.NewServeMux()
	router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, `{"data":{"name":  "a foo b"}}`)
		assert.NoError(t, err)
	})

	request := New(t, router)

	request.Get("/api").
		Expect().
		StatusCode(http.StatusOK).
		JsonPath("$.data.name", Contains("foo"))
}

func Test_NoEmptyMatcher(t *testing.T) {
	router := http.NewServeMux()
	router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, `{"data":{"name":  "a"}}`)
		assert.NoError(t, err)
	})

	request := New(t, router)

	request.Get("/api").
		Expect().
		StatusCode(http.StatusOK).
		JsonPath("$.data.name", NotEmpty())
}
