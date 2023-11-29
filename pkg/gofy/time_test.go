// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gofy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustParseTime(t *testing.T) {
	now := MustParseTime("2022-01-01T01:01:01+02:00")

	assert.NotNil(t, now)
}
