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

type CipherArgon2id struct {
	key []byte
}

func NewCipherArgon2id(password, salt string) *CipherArgon2id {
	key := argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32)
	return &CipherArgon2id{key: key}
}

func (c CipherArgon2id) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nil, plaintext, nil)

	encoded := base64.StdEncoding.AppendEncode(nil, ciphertext)
	return encoded, nil
}

func (c CipherArgon2id) Decrypt(encoded []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.AppendDecode(nil, encoded)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext: %w", err)
	}

	if len(ciphertext) < aes.BlockSize+1 {
		return nil, fmt.Errorf("ciphertext too short")
	}

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
