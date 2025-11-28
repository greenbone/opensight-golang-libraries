// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package httpassert

import (
	"fmt"
	"reflect"
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
		targetVal := reflect.ValueOf(ptr)
		if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
			assert.Fail(t, "ExtractTo requires a non-nil pointer")
			return nil
		}

		if actual == nil {
			assert.Fail(t, "ExtractTo actual value is nil")
			return nil
		}

		outVal := reflect.ValueOf(actual)
		targetType := targetVal.Elem().Type()

		// Direct assign / convert for simple values (string, int, etc.)
		if outVal.Type().AssignableTo(targetType) {
			targetVal.Elem().Set(outVal)
			return ptr
		}

		if outVal.Type().ConvertibleTo(targetType) {
			targetVal.Elem().Set(outVal.Convert(targetType))
			return ptr
		}

		// Special handling for slices, e.g. []interface{} -> []string
		if outVal.Kind() == reflect.Slice && targetType.Kind() == reflect.Slice {
			elemType := targetType.Elem()
			n := outVal.Len()
			dst := reflect.MakeSlice(targetType, n, n)

			for i := 0; i < n; i++ {
				src := outVal.Index(i)

				// If it's interface{}, unwrap to the underlying concrete value.
				if src.Kind() == reflect.Interface && !src.IsNil() {
					src = reflect.ValueOf(src.Interface())
				}

				if src.Type().AssignableTo(elemType) {
					dst.Index(i).Set(src)
					continue
				}

				if src.Type().ConvertibleTo(elemType) {
					dst.Index(i).Set(src.Convert(elemType))
					continue
				}

				assert.Fail(t,
					fmt.Sprintf("ExtractTo slice element type mismatch at index %d: cannot assign %v to %v",
						i, src.Type(), elemType))
				return nil
			}

			targetVal.Elem().Set(dst)
			return ptr
		}

		assert.Fail(t,
			fmt.Sprintf("ExtractTo type mismatch: cannot assign %v to %v",
				outVal.Type(), targetType))
		return nil
	}
}
