// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	validationError := NewValidationError("x:%d y:%s", 1, "one")
	require.Error(t, validationError)
	assert.Equal(t, "x:1 y:one", validationError.Error())
}

func TestInvalidFilterFieldError(t *testing.T) {
	invalidFilterFieldError := NewInvalidFilterFieldError("x:%d y:%s", 1, "one")
	require.Error(t, invalidFilterFieldError)
	assert.Equal(t, "x:1 y:one", invalidFilterFieldError.Error())
}
