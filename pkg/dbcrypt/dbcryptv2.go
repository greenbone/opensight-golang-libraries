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

type DBCryptV2 struct {
	encryptionCipher  dbCipher
	decryptionCiphers map[string]dbCipher
}

// New creates a new instance of DBCryptV2 based on the provided Config.
func New(conf Config) (*DBCryptV2, error) {
	if conf.Password == "" {
		return nil, errors.New("db password is empty")
	}
	if conf.PasswordSalt == "" {
		return nil, errors.New("db password salt is empty")
	}
	if len(conf.PasswordSalt) < 32 {
		return nil, errors.New("db password salt is too short")
	}
	c := &DBCryptV2{}
	if err := c.registerCiphers(conf); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *DBCryptV2) registerCiphers(conf Config) error {
	c.decryptionCiphers = map[string]dbCipher{}
	for _, fn := range dbCiphers {
		cipher, err := fn(conf)
		if err != nil {
			return err
		}
		c.decryptionCiphers[cipher.Prefix()] = cipher
	}

	encryptionCipherPrefix := conf.Version
	if encryptionCipherPrefix == "" {
		encryptionCipherPrefix = defaultCipherPrefix
	}
	cipher := c.decryptionCiphers[encryptionCipherPrefix]
	if cipher == nil {
		return fmt.Errorf("invalid db cipher version %q", conf.Version)
	}
	c.encryptionCipher = cipher
	return nil
}

func (c *DBCryptV2) findDecryptionCipher(ciphertextWithPrefix []byte) (dbCipher, []byte, []byte) {
	i := bytes.Index(ciphertextWithPrefix, []byte(prefixSeparator))
	if i < 0 {
		return nil, nil, ciphertextWithPrefix
	}
	prefix, ciphertext := ciphertextWithPrefix[:i], ciphertextWithPrefix[i+len(prefixSeparator):]
	return c.decryptionCiphers[string(prefix)], prefix, ciphertext
}

func (c *DBCryptV2) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := c.encryptionCipher.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	return append([]byte(c.encryptionCipher.Prefix()+prefixSeparator), ciphertext...), nil
}

func (c *DBCryptV2) Decrypt(ciphertextWithPrefix []byte) ([]byte, error) {
	cipher, prefix, ciphertext := c.findDecryptionCipher(ciphertextWithPrefix)
	if len(prefix) == 0 {
		return nil, errors.New("invalid encrypted value format")
	}
	if cipher == nil {
		return nil, errors.New("unknown encrypted value format")
	}
	plaintext, err := cipher.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
