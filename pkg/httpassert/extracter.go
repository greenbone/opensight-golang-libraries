// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package httpassert

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Extractor func(t *testing.T, actual any) any

// ExtractTo sets the value read from JSONPath into the given pointer variable.
// Example:
//
//	var id string
//	request.Expect().JsonPath("$.data.id", httpassert.ExtractTo(&id))
func ExtractTo(ptr any) Extractor {
	return func(t *testing.T, actual any) any {
		return extractInto(t, ptr, actual, nil)
	}
}

func ExtractRegexTo(value string, ptr any) Extractor {
	return func(t *testing.T, actual any) any {
		return extractInto(t, ptr, actual, func(t *testing.T, v any, dstType reflect.Type) (any, bool) {
			s, ok := v.(string)
			if !ok {
				assert.Fail(t, fmt.Sprintf("ExtractRegexTo expects actual to be string, got %T", v))
				return nil, false
			}

			re := regexp.MustCompile(value)

			m := re.FindStringSubmatch(s)
			if m == nil {
				assert.Fail(t, fmt.Sprintf("ExtractRegexTo no match for %q in %q", re.String(), s))
				return nil, false
			}

			// m[0] full match, m[1:] groups
			groups := m[1:]
			if len(groups) == 0 {
				groups = []string{m[0]}
			}

			if dstType.Kind() == reflect.Slice {
				return groups, true
			}
			return groups[0], true
		})
	}
}

func extractInto(
	t *testing.T,
	ptr any,
	actual any,
	preprocess func(t *testing.T, actual any, dstType reflect.Type) (any, bool),
) any {
	target := reflect.ValueOf(ptr)
	if target.Kind() != reflect.Ptr || target.IsNil() {
		assert.Fail(t, "ExtractTo requires a non-nil pointer")
		return nil
	}
	if actual == nil {
		assert.Fail(t, "ExtractTo actual value is nil")
		return nil
	}

	dst := target.Elem()
	dstType := dst.Type()

	if preprocess != nil {
		var ok bool
		actual, ok = preprocess(t, actual, dstType)
		if !ok {
			return nil
		}
		if actual == nil {
			assert.Fail(t, "ExtractTo actual value is nil")
			return nil
		}
	}

	src := reflect.ValueOf(actual)

	// unwrap interface{}
	if src.IsValid() && src.Kind() == reflect.Interface && !src.IsNil() {
		src = reflect.ValueOf(src.Interface())
	}

	// direct assign / convert
	if src.IsValid() && src.Type().AssignableTo(dstType) {
		dst.Set(src)
		return ptr
	}
	if src.IsValid() && src.Type().ConvertibleTo(dstType) {
		dst.Set(src.Convert(dstType))
		return ptr
	}

	// slice handling: e.g. []interface{} -> []string
	if src.IsValid() && src.Kind() == reflect.Slice && dstType.Kind() == reflect.Slice {
		elemType := dstType.Elem()
		n := src.Len()
		out := reflect.MakeSlice(dstType, n, n)

		for i := 0; i < n; i++ {
			s := src.Index(i)

			// unwrap interface{} elements
			if s.Kind() == reflect.Interface && !s.IsNil() {
				s = reflect.ValueOf(s.Interface())
			}

			if s.Type().AssignableTo(elemType) {
				out.Index(i).Set(s)
				continue
			}
			if s.Type().ConvertibleTo(elemType) {
				out.Index(i).Set(s.Convert(elemType))
				continue
			}

			assert.Fail(t, fmt.Sprintf(
				"ExtractTo slice element type mismatch at index %d: cannot assign %v to %v",
				i, s.Type(), elemType,
			))
			return nil
		}

		dst.Set(out)
		return ptr
	}

	srcType := any("<invalid>")
	if src.IsValid() {
		srcType = src.Type()
	}
	assert.Fail(t, fmt.Sprintf("ExtractTo type mismatch: cannot assign %v to %v", srcType, dstType))
	return nil
}
