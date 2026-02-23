// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package httpassert

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type Matcher func(t *testing.T, actual any) bool

// HasSize checks the length of arrays, maps, or strings.
// Example: JsonPath("$.data", httpassert.HasSize(11))
func HasSize(e int) Matcher {
	return func(t *testing.T, actual any) bool {
		var size int

		switch v := actual.(type) {
		case []any:
			size = len(v)
		case map[string]any:
			size = len(v)
		case string:
			size = len(v)
		default:
			assert.Fail(t, fmt.Sprintf("HasSize: unsupported type %T", actual))
			return false
		}

		valid := assert.Equal(t, e, size,
			fmt.Sprintf("HasSize: expected size %d, got %d", e, size))
		return valid
	}
}

// Contains checks if a string contains the value
// Example: JsonPath("$.data.name", httpassert.Contains("foo"))
func Contains(v string) Matcher {
	return func(t *testing.T, value any) bool {
		valid := assert.Contains(t, value, v)
		return valid
	}
}

// Regex checks if a string matches the given regular expression
// Example: JsonPath("$.data.name", httpassert.Regex("^foo.*bar$"))
func Regex(expr string) Matcher {
	re := regexp.MustCompile(expr)

	return func(t *testing.T, value any) bool {
		return assert.Regexp(t, re, value)
	}
}

// NotEmpty checks if a string is not empty
// Example: JsonPath("$.data.name", httpassert.NotEmpty())
func NotEmpty() Matcher {
	return func(t *testing.T, value any) bool {
		str, ok := value.(string)
		if !ok {
			return assert.Fail(t, "value is not a string")
		}

		return assert.NotEmpty(t, str)
	}
}

// IsUUID checks if a string is a UUID
// Example: JsonPath("$.id", httpassert.IsUUID())
func IsUUID() Matcher {
	return func(t *testing.T, value any) bool {
		t.Helper()

		str, ok := value.(string)
		if !ok {
			return assert.Fail(t, "value is not a string")
		}

		_, err := uuid.Parse(str)
		if err != nil {
			return assert.Fail(t, "value is not a valid UUID", "'%s': %v", str, err)
		}

		return true
	}
}
