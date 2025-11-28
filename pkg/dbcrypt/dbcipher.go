// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"bytes"
	"errors"
	"fmt"
)

const prefixSeparator = ":"

// Config encapsulates configuration for DBCipher.
type Config struct {
	// Default version of the cryptographic algorithm. Useful for testing older historical implementations. Leave empty to use the most recent version.
	//
	// Supported values:
	// - "": use latest version of the cryptographic algorithm (recommended).
	// - "v2": use v2 version of the cryptographic algorithm.
	// - "v1": use v1 version of the cryptographic algorithm.
	//
	// See cipher_spec.go for all versions
	Version string

	// Contains the password used to derive encryption key
	Password string

	// Contains the salt for increasing password entropy
	PasswordSalt string
}

// Validate validates the provided config.
func (conf Config) Validate() error {
	if conf.Password == "" {
		return errors.New("db password is empty")
	}
	if conf.PasswordSalt == "" {
		return errors.New("db password salt is empty")
	}
	if len(conf.PasswordSalt) < 32 {
		return errors.New("db password salt is too short")
	}
	return nil
}

// DBCipher is cipher designed to perform validated encryption and decryption on database values.
type DBCipher struct {
	encryptionCipherSpec *cipherSpec
	ciphersSpec          *ciphersSpec
}

// NewDBCipher creates a new instance of DBCipher based on the provided Config.
func NewDBCipher(conf Config) (*DBCipher, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	spec, err := newCiphersSpec(conf)
	if err != nil {
		return nil, fmt.Errorf("error creating crypto ciphers spec: %w", err)
	}

	encryptionVersion := conf.Version
	if encryptionVersion == "" {
		encryptionVersion = spec.DefaultVersion
	}

	encryptionCipherSpec, err := spec.GetByVersion(encryptionVersion)
	if err != nil {
		return nil, fmt.Errorf("could not get encryption cipher by version: %w", err)
	}

	c := &DBCipher{
		encryptionCipherSpec: encryptionCipherSpec,
		ciphersSpec:          spec,
	}
	return c, nil
}

// Encrypt encrypts the provided bytes with DBCipher.
func (c *DBCipher) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := c.encryptionCipherSpec.Cipher.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	ciphertextWithPrefix := bytes.NewBuffer(nil)
	ciphertextWithPrefix.WriteString(c.encryptionCipherSpec.Prefix)
	ciphertextWithPrefix.WriteString(prefixSeparator)
	ciphertextWithPrefix.Write(ciphertext)
	return ciphertextWithPrefix.Bytes(), nil
}

// Decrypt decrypts the provided bytes with DBCipher.
func (c *DBCipher) Decrypt(ciphertextWithPrefix []byte) ([]byte, error) {
	if len(ciphertextWithPrefix) == 0 {
		return nil, nil
	}
	prefix, ciphertext, hasSeparator := bytes.Cut(ciphertextWithPrefix, []byte(prefixSeparator))
	if !hasSeparator {
		return nil, errors.New("invalid encrypted value format")
	}
	decryptionCipherSpec, err := c.ciphersSpec.GetByPrefix(string(prefix))
	if err != nil {
		return nil, fmt.Errorf("unknown encrypted value format: %w", err)
	}
	plaintext, err := decryptionCipherSpec.Cipher.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
