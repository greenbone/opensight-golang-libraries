// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type CipherAES struct {
	key []byte
}

func NewCipherAES(password, salt string) *CipherAES {
	key := []byte(password + salt)[:32]
	return &CipherAES{key: key}
}

func (c CipherAES) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, nil)
	ciphertextWithIv := append(iv, ciphertext...)

	encoded := hex.AppendEncode(nil, ciphertextWithIv)
	return encoded, nil
}

func (c CipherAES) Decrypt(encoded []byte) ([]byte, error) {
	ciphertextWithIv, err := hex.AppendDecode(nil, encoded)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext: %w", err)
	}

	if len(ciphertextWithIv) < aes.BlockSize+1 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	iv := ciphertextWithIv[:gcm.NonceSize()]
	ciphertext := ciphertextWithIv[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting ciphertext: %w", err)
	}

	return plaintext, nil
}
