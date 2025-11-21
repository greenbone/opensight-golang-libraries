// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"bytes"
	"errors"
	"fmt"
)

type Crypter interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type CryptoManager struct {
	cipherManager    *CipherManager
	encryptionCipher *Cipher
	prefixSeparator  string
}

// NewCryptoManager creates a new instance of CryptoManager based on the provided Config.
func NewCryptoManager(conf Config) (*CryptoManager, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	manager, err := NewCipherManager(conf)
	if err != nil {
		return nil, fmt.Errorf("error creating crypto manager: %w", err)
	}

	version := conf.Version
	if version == "" {
		version = manager.GetDefaultVersion()
	}

	cipher, err := manager.GetByVersion(version)
	if err != nil {
		return nil, fmt.Errorf("could not get cipher by version: %w", err)
	}

	c := &CryptoManager{
		prefixSeparator:  ":", // this can never change, otherwise existing data will break
		encryptionCipher: cipher,
		cipherManager:    manager,
	}

	return c, nil
}

// Encrypt encrypts the provided bytes and adds a prefix for the used implementation
func (c *CryptoManager) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := c.encryptionCipher.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}

	ciphertextWithPrefix := bytes.NewBuffer(nil)
	ciphertextWithPrefix.WriteString(c.encryptionCipher.Prefix)
	ciphertextWithPrefix.WriteString(c.prefixSeparator)
	ciphertextWithPrefix.Write(ciphertext)
	return ciphertextWithPrefix.Bytes(), nil
}

// Decrypt decrypts the provided bytes that are prefix with the implementation
func (c *CryptoManager) Decrypt(ciphertextWithPrefix []byte) ([]byte, error) {
	prefix, ciphertext, hasSeparator := bytes.Cut(ciphertextWithPrefix, []byte(c.prefixSeparator))
	if !hasSeparator {
		return nil, errors.New("invalid encrypted value format")
	}

	cipher, err := c.cipherManager.GetByPrefix(string(prefix))
	if err != nil {
		return nil, fmt.Errorf("unknown encrypted value format: %w", err)
	}

	plaintext, err := cipher.Decrypt(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted value: %w", err)
	}

	return plaintext, nil
}
