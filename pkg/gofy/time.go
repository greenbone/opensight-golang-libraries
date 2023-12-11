// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gofy

import "time"

// MustParseTime parses the given string `value` as a time using the RFC3339 format.
// If the parsing fails, it will panic with the error returned by `time.Parse`.
// It returns the parsed time as a `time.Time` value.
// Example Usage:
//
//	timeValue := MustParseTime("2022-01-01T01:01:01+02:00")
//	fmt.Println(timeValue)
//
// Output:
//
//	2022-01-01 01:01:01 +0200 +0200
func MustParseTime(value string) time.Time {
	timeValue, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}

	return timeValue
}
