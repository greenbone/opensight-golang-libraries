// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package slices

// ContainsLambda checks if an element in the given slice satisfies the provided filter function.
// It returns true if at least one element satisfies the filter function, otherwise it returns false.
// The filter function takes an element of type T as input and returns a bool.
// The elements slice is of type []T.
func ContainsLambda[T any](elements []T, filterFunction func(element T) bool) bool {
	for _, element := range elements {
		if filterFunction(element) {
			return true
		}
	}
	return false
}

// Contains is a generic function that checks if a given value exists in a slice of comparable elements.
// It takes in a slice of elements and a value, and returns true if the value is found in the slice, false otherwise.
func Contains[T comparable](elements []T, value T) bool {
	for _, element := range elements {
		if element == value {
			return true
		}
	}
	return false
}
