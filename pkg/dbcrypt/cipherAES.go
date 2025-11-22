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

type cipherAES struct {
	key []byte
}

func truncate(value string, length int) string {
	if len(value) <= length {
		return value
	}
	return value[:length]
}

func newCipherAES(password, salt string) *cipherAES {
	key := []byte(truncate(password+salt, 32))
	return &cipherAES{key: key}
}

func (c cipherAES) Encrypt(plaintext []byte) ([]byte, error) {
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

	return ciphertextWithIv, nil
}

func (c cipherAES) Decrypt(ciphertextWithIv []byte) ([]byte, error) {
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

type CipherAesHex struct {
	encrypter *cipherAES
}

func NewCipherAesHex(password, salt string) *CipherAesHex {
	return &CipherAesHex{
		encrypter: newCipherAES(password, salt),
	}
}

func (c CipherAesHex) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertextWithIv, err := c.encrypter.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}

	return hex.AppendEncode(nil, ciphertextWithIv), nil
}

func (c CipherAesHex) Decrypt(encoded []byte) ([]byte, error) {
	ciphertextWithIv, err := hex.AppendDecode(nil, encoded)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext: %w", err)
	}

	return c.encrypter.Decrypt(ciphertextWithIv)
}
