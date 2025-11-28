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

func TestExtractTo(t *testing.T) {
	router := http.NewServeMux()
	router.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, `{"data":{
			"id":"abc-123", 
			"ids":["a", "b"],
			"number":123 ,
			"numbers":[1,2]
   		}}`)
		assert.NoError(t, err)
	})

	request := New(t, router)

	t.Run("string", func(t *testing.T) {
		var id string
		request.Post("/data").
			Content(`{"name":"appliance1"}`).
			Expect().
			StatusCode(http.StatusOK).
			JsonPath("$.data.id", ExtractTo(&id))

		require.NotEmpty(t, id)
		require.Equal(t, "abc-123", id)
	})

	t.Run("string list", func(t *testing.T) {
		var ids []string
		request.Post("/data").
			Content(`{"name":"appliance1"}`).
			Expect().
			StatusCode(http.StatusOK).
			JsonPath("$.data.ids", ExtractTo(&ids))

		require.NotEmpty(t, ids)
		require.Equal(t, []string{"a", "b"}, ids)
	})

	t.Run("number", func(t *testing.T) {
		var number int
		request.Post("/data").
			Content(`{"name":"appliance1"}`).
			Expect().
			StatusCode(http.StatusOK).
			JsonPath("$.data.number", ExtractTo(&number))

		require.NotEmpty(t, number)
		require.Equal(t, 123, number)
	})

	t.Run("number list", func(t *testing.T) {
		var numbers []int
		request.Post("/data").
			Content(`{"name":"appliance1"}`).
			Expect().
			StatusCode(http.StatusOK).
			JsonPath("$.data.numbers", ExtractTo(&numbers))

		require.NotEmpty(t, numbers)
		require.Equal(t, []int{1, 2}, numbers)
	})
}
