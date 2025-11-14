// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
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
	"reflect"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"github.com/greenbone/opensight-golang-libraries/pkg/dbcrypt/config"
)

const (
	prefix    = "ENC:"
	prefixLen = len(prefix)
)

type DBCrypt[T any] struct {
	config atomic.Pointer[config.CryptoConfig]
}

// New creates a new instance of DBCrypt that will perform cryptographic operations based on the provided CryptoConfig.
func New[T any](config config.CryptoConfig) *DBCrypt[T] {
	d := &DBCrypt[T]{}
	d.config.Store(&config)
	return d
}

func (d *DBCrypt[T]) loadKey() []byte {
	configPtr := d.config.Load()
	if configPtr == nil {
		conf, err := config.Read()
		if err != nil {
			log.Fatal().Err(err).Msg("crypto config is invalid")
		}
		d.config.CompareAndSwap(nil, &conf)
		configPtr = d.config.Load()
	}
	// TODO: Use proper KDF instead of truncating the password (to maintain backward compatibility use encrypted values prefixes to determine the method)
	key := []byte(configPtr.ReportEncryptionV1Password + configPtr.ReportEncryptionV1Salt)[:32] // Truncate key to 32 bytes
	return key
}

// EncryptStruct encrypts all fields of a struct that are tagged with `encrypt:"true"`
func (d *DBCrypt[T]) EncryptStruct(data *T) error {
	key := d.loadKey()
	value := reflect.ValueOf(data).Elem()
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := valueType.Field(i)
		if encrypt, ok := fieldType.Tag.Lookup("encrypt"); ok && encrypt == "true" {
			plaintext := fmt.Sprintf("%v", field.Interface())
			if len(plaintext) > prefixLen && plaintext[:prefixLen] == prefix {
				// already encrypted goto next field
				continue
			}
			ciphertext, err := Encrypt(plaintext, key)
			if err != nil {
				return err
			}
			field.SetString(ciphertext)
		}
	}
	return nil
}

// DecryptStruct decrypts all fields of a struct that are tagged with `encrypt:"true"`
func (d *DBCrypt[T]) DecryptStruct(data *T) error {
	key := d.loadKey()
	value := reflect.ValueOf(data).Elem()
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := valueType.Field(i)
		if encrypt, ok := fieldType.Tag.Lookup("encrypt"); ok && encrypt == "true" {
			plaintext, err := Decrypt(field.String(), key)
			if err != nil {
				return err
			}
			field.SetString(plaintext)
		}
	}
	return nil
}

func Encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)

	encoded := hex.EncodeToString(append(iv, ciphertext...))
	return prefix + encoded, nil
}

func Decrypt(encrypted string, key []byte) (string, error) {
	if len(encrypted) <= prefixLen || encrypted[:prefixLen] != prefix {
		return "", fmt.Errorf("invalid encrypted value format")
	}

	encodedCiphertext := encrypted[4:]

	ciphertext, err := hex.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("error decoding ciphertext: %w", err)
	}

	if len(ciphertext) < aes.BlockSize+1 {
		return "", fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	iv := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("error decrypting ciphertext: %w", err)
	}

	return string(plaintext), nil
}
