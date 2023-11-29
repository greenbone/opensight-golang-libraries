// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package test_folder

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestFolder interface {
	// GetContent get content
	GetContent(t *testing.T, path string) string

	// GetReader get content reader
	GetReader(t *testing.T, path string) io.Reader
}

type testFolder struct{}

func (f *testFolder) GetContent(t *testing.T, filename string) string {
	t.Helper()

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	return string(content)
}

func (f *testFolder) GetReader(t *testing.T, filename string) io.Reader {
	t.Helper()

	reader, err := os.Open(filename)
	require.NoError(t, err)

	return reader
}

func NewTestFolder() TestFolder {
	return &testFolder{}
}
