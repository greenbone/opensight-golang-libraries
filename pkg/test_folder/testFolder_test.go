// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package test_folder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestFolder(t *testing.T) {
	var testFolder TestFolder

	setup := func(t *testing.T) {
		testFolder = NewTestFolder()
	}

	t.Run("shouldReturnContentFromFile", func(t *testing.T) {
		setup(t)

		content := testFolder.GetContent(t, "testFolder_test.go")

		assert.NotEmpty(t, content)
	})
}
