// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gofy

import "time"

func MustParseTime(value string) time.Time {
	timeValue, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}

	return timeValue
}
