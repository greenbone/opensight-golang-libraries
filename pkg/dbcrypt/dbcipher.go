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
	// - use <empty> for v2 version of the cryptographic algorithm
	// - use "v2" for v2 version of the cryptographic algorithm
	// - use "v1" for v1 version of the cryptographic algorithm
	Version string

	// Contains the password used deriving encryption key
	Password string

	// Contains the salt for increasing password entropy
	PasswordSalt string
}

// DBCipher is cipher designed to perform validated encryption and decryption on database values.
type DBCipher struct {
	encryptionCipher  dbCipher
	decryptionCiphers map[string]dbCipher
}

// NewDBCipher creates a new instance of DBCipher based on the provided Config.
func NewDBCipher(conf Config) (*DBCipher, error) {
	if conf.Password == "" {
		return nil, errors.New("db password is empty")
	}
	if conf.PasswordSalt == "" {
		return nil, errors.New("db password salt is empty")
	}
	if len(conf.PasswordSalt) < 32 {
		return nil, errors.New("db password salt is too short")
	}
	c := &DBCipher{}
	if err := c.registerCiphers(conf); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *DBCipher) registerCiphers(conf Config) error {
	v2, err := newDbCipherV2(conf)
	if err != nil {
		return err
	}
	v1, err := newDbCipherV1(conf)
	if err != nil {
		return err
	}

	c.decryptionCiphers = map[string]dbCipher{
		v2.Prefix(): v2,
		v1.Prefix(): v1,
	}
	switch conf.Version {
	case "", "v2":
		c.encryptionCipher = v2
	case "v1":
		c.encryptionCipher = v1
	default:
		return fmt.Errorf("invalid db cipher version %q", conf.Version)
	}
	return nil
}

// Encrypt encrypts the provided bytes with DBCipher.
func (c *DBCipher) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := c.encryptionCipher.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	return append([]byte(c.encryptionCipher.Prefix()+prefixSeparator), ciphertext...), nil
}

// Decrypt decrypts the provided bytes with DBCipher.
func (c *DBCipher) Decrypt(ciphertextWithPrefix []byte) ([]byte, error) {
	prefix, ciphertext, hasSeparator := bytes.Cut(ciphertextWithPrefix, []byte(prefixSeparator))
	if !hasSeparator {
		return nil, errors.New("invalid encrypted value format")
	}
	cipher := c.decryptionCiphers[string(prefix)]
	if cipher == nil {
		return nil, errors.New("unknown encrypted value format")
	}
	plaintext, err := cipher.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
