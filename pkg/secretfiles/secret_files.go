// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package secretfiles eases the parsing of secret files into a string.
// This is a common scenario when working with docker secrets, where the
// path to the secret is usually stored in an environment variable <SECRET>_FILE
// containing a path to a secret stored in the container filesystem.
package secretfiles

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ReadSecret reads the file path from the given environment variable
// and writes the content of this file into the target.
// Surrounding whitespaces are truncated.
func ReadSecret(envVar string, target *string) error {
	if target == nil {
		return errors.New("can't set target value, as it is nil")
	}
	path := os.Getenv(envVar)
	if path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read secret from file: %w", err)
		}
		secret := strings.TrimSpace(string(content))
		if secret == "" {
			return fmt.Errorf("secret in file %s is empty", path)
		}
		*target = strings.TrimSpace(string(content))
	}
	return nil
}
