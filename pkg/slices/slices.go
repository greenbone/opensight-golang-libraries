// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package slices

func ContainsLambda[T any](elements []T, filterFunction func(element T) bool) bool {
	for _, element := range elements {
		if filterFunction(element) {
			return true
		}
	}
	return false
}

func Contains[T comparable](elements []T, value T) bool {
	for _, element := range elements {
		if element == value {
			return true
		}
	}
	return false
}
