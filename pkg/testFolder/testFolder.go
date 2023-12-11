// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package testFolder

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFolder is an interface that provides methods for accessing content and reading files from a test folder.
// GetContent retrieves the content of a file located at a specified path.
// It takes a *testing.T object and a string path as parameters.
// It returns the content of the file as a string.
type TestFolder interface {
	// GetContent get content
	GetContent(t *testing.T, path string) string

	// GetReader get content reader
	GetReader(t *testing.T, path string) io.Reader
}

// testFolder represents a test folder implementation.
type testFolder struct{}

// GetContent retrieves the content of a file with the given filename.
// It reads the file using `os.ReadFile` and returns the content as a string.
// If any error occurs during file reading, it fails the test and throws an error.
// The `t` parameter is a testing.T instance used for error reporting and test assertion.
// Example usage: content := testFolder.GetContent(t, "file.txt")
// where `testFolder` is an instance of the `testFolder` type and `file.txt` is the name of the file to read.
func (f *testFolder) GetContent(t *testing.T, filename string) string {
	t.Helper()

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	return string(content)
}

// GetReader returns an `io.Reader` for the given `filename`. It opens the file for reading and returns a reader interface. It may also return an error if there was a problem opening
func (f *testFolder) GetReader(t *testing.T, filename string) io.Reader {
	t.Helper()

	reader, err := os.Open(filename)
	require.NoError(t, err)

	return reader
}

// NewTestFolder returns a new instance of the TestFolder interface. It creates an instance of the testFolder struct and returns it as a TestFolder.
func NewTestFolder() TestFolder {
	return &testFolder{}
}
