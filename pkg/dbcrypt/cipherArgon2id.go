// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type cipherArgon2id struct {
	key []byte
}

func newCipherArgon2id(password, salt string) *cipherArgon2id {
	key := argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32)
	return &cipherArgon2id{key: key}
}

func (c cipherArgon2id) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nil, nil, plaintext, nil), nil
}

func (c cipherArgon2id) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nil, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting ciphertext: %w", err)
	}

	return plaintext, nil
}

type CipherArgon2idBase64 struct {
	encrypter *cipherArgon2id
}

func NewCipherArgon2idBase64(password, salt string) *CipherArgon2idBase64 {
	return &CipherArgon2idBase64{
		encrypter: newCipherArgon2id(password, salt),
	}
}

func (c CipherArgon2idBase64) Encrypt(plaintext []byte) ([]byte, error) {
	rawCiphertext, err := c.encrypter.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.AppendEncode(nil, rawCiphertext), nil
}

func (c CipherArgon2idBase64) Decrypt(encoded []byte) ([]byte, error) {
	rawCiphertext, err := base64.StdEncoding.AppendDecode(nil, encoded)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext: %w", err)
	}

	return c.encrypter.Decrypt(rawCiphertext)
}
