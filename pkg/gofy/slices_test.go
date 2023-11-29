// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gofy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsLambda(t *testing.T) {
	t.Run("shouldReturnTrueWhenFound", func(t *testing.T) {
		output := ContainsLambda([]string{"a", "b"}, func(value string) bool {
			return value == "a"
		})

		assert.True(t, output)
	})

	t.Run("shouldReturnFalseWhenNotFound", func(t *testing.T) {
		output := ContainsLambda([]string{"a", "b"}, func(value string) bool {
			return value == "nope"
		})

		assert.False(t, output)
	})
}

func TestContains(t *testing.T) {
	t.Run("shouldReturnTrueWhenFound", func(t *testing.T) {
		output := Contains([]string{"a", "b"}, "a")

		assert.True(t, output)
	})

	t.Run("shouldReturnFalseWhenNotFound", func(t *testing.T) {
		output := Contains([]string{"a", "b"}, "nope")

		assert.False(t, output)
	})
}
