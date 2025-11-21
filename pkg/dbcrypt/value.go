package dbcrypt

import (
	"database/sql/driver"
	"errors"
)

// EncryptedString is a wrapper around string that indicates that the value should be encrypted wile stored.
type EncryptedString struct {
	encrypted string
	decrypted string
}

// NewEncryptedString creates a new EncryptedString based on plaintext data. Returned value, until encrypted, will miss associated ciphertext value.
func NewEncryptedString(dec string) *EncryptedString {
	return &EncryptedString{decrypted: dec}
}

// DecryptEncryptedString creates a new EncryptedString based on ciphertext data. It automatically decrypts it using the provided DBCipher, so both plaintext and ciphertext values are available.
func DecryptEncryptedString(c *DBCipher, enc string) (*EncryptedString, error) {
	es := &EncryptedString{encrypted: enc}
	if err := es.decrypt(c); err != nil {
		return nil, err
	}
	return es, nil
}

// Scan unmarshal encrypted stored value into EncryptedString.
func (es *EncryptedString) Scan(v any) error {
	enc, ok := v.(string)
	if !ok {
		return errors.New("failed to unmarshal encrypted string value")
	}
	es.encrypted, es.decrypted = enc, ""
	return nil
}

// Value returns encrypted value for storing.
func (es EncryptedString) Value() (driver.Value, error) {
	enc, ok := es.Encrypted()
	if !ok {
		return nil, errors.New("cannot store string value: encryption required")
	}
	if enc == "" {
		return nil, nil
	}
	return enc, nil
}

// Encrypted returns ciphertext (encrypted) value of EncryptedString and true, if encrypted value is available. Otherwise it return an empty string and false.
func (es *EncryptedString) Encrypted() (string, bool) {
	if es == nil {
		return "", true
	}
	has := es.encrypted != "" || es.decrypted == ""
	return es.encrypted, has
}

// Encrypt generates a new ciphertext value based on plaintext value using the provided DBCipher.
func (es *EncryptedString) Encrypt(c *DBCipher) error {
	enc, err := c.Encrypt([]byte(es.decrypted))
	if err != nil {
		return err
	}
	es.encrypted = string(enc)
	return nil
}

// ClearEncrypted removes associated encrypted value.
func (es *EncryptedString) ClearEncrypted() {
	es.encrypted = ""
}

func (es *EncryptedString) decrypt(c *DBCipher) error {
	enc, ok := es.Encrypted()
	if !ok || enc == "" {
		return nil
	}
	dec, err := c.Decrypt([]byte(enc))
	if err != nil {
		return err
	}
	es.decrypted = string(dec)
	return nil
}

// Get returns plaintext (decrypted) value of EncryptedString.
func (es *EncryptedString) Get() string {
	if es == nil {
		return ""
	}
	return es.decrypted
}

// Set sets plaintext (decrypted) value of EncryptedString.
func (es *EncryptedString) Set(to string) {
	es.encrypted, es.decrypted = "", to
}
