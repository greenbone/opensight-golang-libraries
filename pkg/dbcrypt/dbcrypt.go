package dbcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"

	"github.com/greenbone/opensight-golang-libraries/pkg/dbcrypt/config"
)

type DBCrypt[T any] struct {
	config config.CryptoConfig
}

func (d *DBCrypt[T]) loadConfig() {
	if d.config == (config.CryptoConfig{}) {
		d.config = config.Read()
	}
}

func (d *DBCrypt[T]) EncryptStruct(data *T) error {
	d.loadConfig()
	value := reflect.ValueOf(data).Elem()
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := valueType.Field(i)
		if encrypt, ok := fieldType.Tag.Lookup("encrypt"); ok && encrypt == "true" {
			plaintext := fmt.Sprintf("%v", field.Interface())
			if len(plaintext) > 4 && plaintext[:4] == "ENC:" {
				// already encrypted goto next field
				continue
			}
			ciphertext, err := d.encrypt(plaintext)
			if err != nil {
				return err
			}
			field.SetString(ciphertext)
			// field.SetString("ENC:" + hex.EncodeToString(ciphertext))
		}
	}
	return nil
}

func (d *DBCrypt[T]) DecryptStruct(data *T) error {
	d.loadConfig()
	value := reflect.ValueOf(data).Elem()
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := valueType.Field(i)
		if encrypt, ok := fieldType.Tag.Lookup("encrypt"); ok && encrypt == "true" {
			plaintext, err := d.decrypt(field.String())
			if err != nil {
				return err
			}
			field.SetString(plaintext)
		}
	}
	return nil
}

func (d *DBCrypt[T]) encrypt(plaintext string) (string, error) {
	key := []byte(d.config.ReportEncryptionV1Password + d.config.ReportEncryptionV1Salt)[:32] // Truncate or pad

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
	return "ENC:" + encoded, nil
}

func (d *DBCrypt[T]) decrypt(encrypted string) (string, error) {
	if len(encrypted) < 5 || encrypted[:4] != "ENC:" {
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

	key := []byte(d.config.ReportEncryptionV1Password + d.config.ReportEncryptionV1Salt)[:32] // Truncate or pad key to 32 bytes

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
