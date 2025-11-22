// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/greenbone/opensight-golang-libraries/pkg/dbcrypt"
	"github.com/stretchr/testify/require"
)

func TestCipherEncryptAndDecrypt(t *testing.T) {
	tests := []struct {
		name   string
		config dbcrypt.Config
		given  string
	}{
		{
			name: "latest/random",
			config: dbcrypt.Config{
				Version:      "",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			given: uuid.NewString(),
		},
		{
			name: "v2/random",
			config: dbcrypt.Config{
				Version:      "v2",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			given: uuid.NewString(),
		},
		{
			name: "v2/empty",
			config: dbcrypt.Config{
				Version:      "v2",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			given: "",
		},
		{
			name: "v2/prefix",
			config: dbcrypt.Config{
				Version:      "v2",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			given: "ENCV2:",
		},
		{
			name: "v1/random",
			config: dbcrypt.Config{
				Version:      "v1",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			given: uuid.NewString(),
		},
		{
			name: "v1/empty",
			config: dbcrypt.Config{
				Version:      "v1",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			given: "",
		},
		{
			name: "v1/prefix",
			config: dbcrypt.Config{
				Version:      "v1",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			given: "ENC:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cryptoService, err := dbcrypt.NewCryptoService(test.config)
			require.NoError(t, err)

			ciphertext, err := cryptoService.Encrypt([]byte(test.given))
			require.NoError(t, err)
			require.NotEqual(t, test.given, string(ciphertext))

			got, err := cryptoService.Decrypt(ciphertext)
			require.NoError(t, err)
			require.Equal(t, test.given, string(got))
		})
	}
}

func TestCipherCreationFailure(t *testing.T) {
	tests := []struct {
		name               string
		config             dbcrypt.Config
		errorShouldContain string
	}{
		{
			name: "unknown-version",
			config: dbcrypt.Config{
				Version:      "unknown",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			errorShouldContain: "could not get cipher by version: version 'unknown' not found",
		},
		{
			name: "empty-password",
			config: dbcrypt.Config{
				Version:      "",
				Password:     "",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			errorShouldContain: "password is empty",
		},
		{
			name: "empty-salt",
			config: dbcrypt.Config{
				Version:      "",
				Password:     "encryption-password",
				PasswordSalt: "",
			},
			errorShouldContain: "salt is empty",
		},
		{
			name: "salt-too-short",
			config: dbcrypt.Config{
				Version:      "",
				Password:     "encryption-password",
				PasswordSalt: "short-salt",
			},
			errorShouldContain: "salt is too short",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := dbcrypt.NewCryptoService(test.config)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.errorShouldContain)
		})
	}
}

func TestHistoricalDataDecryption(t *testing.T) {
	tests := []struct {
		name      string
		config    dbcrypt.Config
		encrypted string
		decrypted string
	}{
		{
			name: "v1/simple",
			config: dbcrypt.Config{
				Version:      "v1",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			encrypted: "ENC:425378be21051852fbbc94dca44e1837039ab68f61b02e02bb731c3de0930b6bdb00",
			decrypted: "FooBar",
		},
		{
			name: "v1/salt-truncation", // "v1" historically uses insecure password-salt truncation, this test checks preservation of this behavior
			config: dbcrypt.Config{
				Version:      "v1",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456789-0123456789",
			},
			encrypted: "ENC:425378be21051852fbbc94dca44e1837039ab68f61b02e02bb731c3de0930b6bdb00",
			decrypted: "FooBar",
		},
		{
			name: "v1/password-truncation", // "v1" historically uses insecure password-salt truncation, this test checks preservation of this behavior
			config: dbcrypt.Config{
				Version:      "v1",
				Password:     "encryption-password-0123456789-0123456789",
				PasswordSalt: "0123456789-0123456789-0123456789",
			},
			encrypted: "ENC:d18d84eb52946f069ee6b967c84657f9d9cf7d89940685ad348a161d2e212f16f5be",
			decrypted: "FooBar",
		},
		{
			name: "v2/simple",
			config: dbcrypt.Config{
				Version:      "v2",
				Password:     "encryption-password",
				PasswordSalt: "encryption-password-salt-0123456",
			},
			encrypted: "ENCV2:gr6MB6TefIXMvKwc0DRkBuApxHiu9tvAqO+FHnEeRgMuqQ==",
			decrypted: "FooBar",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cryptoService, err := dbcrypt.NewCryptoService(test.config)
			require.NoError(t, err)

			got, err := cryptoService.Decrypt([]byte(test.encrypted))
			require.NoError(t, err)
			require.Equal(t, test.decrypted, string(got))
		})
	}
}
