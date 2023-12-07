package dbcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/greenbone/opensight-golang-libraries/pkg/dbcrypt/config"
	"io"
	"reflect"
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
			/*
				ciphertext, err := hex.DecodeString(field.String()[4:])
				if err != nil {
					return err
				}*/

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

/*
// This function encrypts a value using AES encryption with the derived key
func (d *DBCrypt[T]) encryptValue(value string) ([]byte, error) {
	key, genKeyErr := d.deriveEncryptionKey()
	if genKeyErr != nil {
		return nil, genKeyErr
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return ciphertext, nil
}

// This function decrypts a value using AES decryption with the derived key
func (d *DBCrypt[T]) decryptValue(ciphertext []byte) (string, error) {
	key, genKeyErr := d.deriveEncryptionKey()
	if genKeyErr != nil {
		return "", genKeyErr
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

*/

// TODO test without DB
// TODO Test DB beforeUpdate ...
