// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type dbCipher interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type dbCipherGcmAes struct {
	key []byte
}

func newDbCipherGcmAes(key []byte) dbCipher {
	return dbCipherGcmAes{key: key}
}

func newDbCipherGcmAesWithoutKdf(password, passwordSalt string) dbCipher {
	// Historically "v1" uses key truncation to 32 bytes. It needs to be preserved for backward compatibility.
	key := make([]byte, 32)
	copy(key, []byte(password+passwordSalt))
	return newDbCipherGcmAes(key)
}

func newDbCipherGcmAesWithArgon2idKdf(password, passwordSalt string) dbCipher {
	// "v2" uses proper KDF (argon2id) to get the key.
	key := argon2.IDKey([]byte(password), []byte(passwordSalt), 1, 64*1024, 4, 32)
	return newDbCipherGcmAes(key)
}

func (c dbCipherGcmAes) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, fmt.Errorf("error encrypting plaintext: %w", err)
	}

	ciphertext := gcm.Seal(nil, nil, []byte(plaintext), nil)
	return ciphertext, nil
}

func (c dbCipherGcmAes) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, fmt.Errorf("error decrypting ciphertext: %w", err)
	}

	plaintext, err := gcm.Open(nil, nil, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting ciphertext: %w", err)
	}

	return plaintext, nil
}

type dbCipherHexEncode struct {
	impl dbCipher
}

func newDbCipherHexEncode(impl dbCipher) dbCipher {
	return dbCipherHexEncode{impl: impl}
}

func (c dbCipherHexEncode) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := c.impl.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	encoded := hex.AppendEncode(nil, ciphertext)
	return encoded, nil
}

func (c dbCipherHexEncode) Decrypt(encoded []byte) ([]byte, error) {
	ciphertext, err := hex.AppendDecode(nil, encoded)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext: %w", err)
	}
	return c.impl.Decrypt(ciphertext)
}

type dbCipherBase64Encode struct {
	impl dbCipher
}

func newDbCipherBase64Encode(impl dbCipher) dbCipher {
	return dbCipherBase64Encode{impl: impl}
}

func (c dbCipherBase64Encode) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := c.impl.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.AppendEncode(nil, ciphertext)
	return encoded, nil
}

func (c dbCipherBase64Encode) Decrypt(encoded []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.AppendDecode(nil, encoded)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext: %w", err)
	}
	return c.impl.Decrypt(ciphertext)
}
