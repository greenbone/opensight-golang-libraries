// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

type dbCipher interface {
	Prefix() string
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type dbCipherV1 struct {
	key []byte
}

func newDbCipherV1(conf Config) (dbCipher, error) {
	// Historically "v1" uses key truncation to 32 bytes. It needs to be preserved for backward compatibility.
	key := []byte(conf.Password + conf.PasswordSalt)[:32]
	return dbCipherV1{key: key}, nil
}

func (c dbCipherV1) Prefix() string {
	return "ENC"
}

func (c dbCipherV1) Encrypt(plaintext []byte) ([]byte, error) {
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

	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)
	ciphertextWithIv := append(iv, ciphertext...)
	encoded := hex.AppendEncode(nil, ciphertextWithIv)
	return encoded, nil
}

func (c dbCipherV1) Decrypt(encoded []byte) ([]byte, error) {
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

type dbCipherV2 struct {
	key []byte
}

func newDbCipherV2(conf Config) (dbCipher, error) {
	// "v2" uses proper KDF (argon2id) to get the key
	key := argon2.IDKey([]byte(conf.Password), []byte(conf.PasswordSalt), 1, 64*1024, 4, 32)
	return dbCipherV2{key: key}, nil
}

func (c dbCipherV2) Prefix() string {
	return "ENCV2"
}

func (c dbCipherV2) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nil, []byte(plaintext), nil)
	encoded := base64.StdEncoding.AppendEncode(nil, ciphertext)
	return encoded, nil
}

func (c dbCipherV2) Decrypt(encoded []byte) ([]byte, error) {
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
