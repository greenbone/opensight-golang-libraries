package dbcrypt

import (
	"database/sql/driver"
	"errors"
)

type EncryptedString struct {
	encrypted string
	decrypted string
}

func NewEncryptedString(val string) *EncryptedString {
	return &EncryptedString{decrypted: val}
}

func (es *EncryptedString) Scan(v any) error {
	enc, ok := v.(string)
	if !ok {
		return errors.New("failed to unmarshal encrypted string value")
	}
	es.encrypted, es.decrypted = enc, ""
	return nil
}

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

func (es *EncryptedString) Encrypted() (string, bool) {
	if es == nil {
		return "", true
	}
	has := es.encrypted != "" || es.decrypted == ""
	return es.encrypted, has
}

func (es *EncryptedString) Encrypt(c *DBCryptV2) error {
	enc, err := EncryptString(c, es.decrypted)
	if err != nil {
		return err
	}
	es.encrypted = enc
	return nil
}

func (es *EncryptedString) ClearEncrypted() {
	es.encrypted = ""
}

func (es *EncryptedString) decrypt(c *DBCryptV2) error {
	enc, ok := es.Encrypted()
	if !ok || enc == "" {
		return nil
	}
	dec, err := DecryptString(c, enc)
	if err != nil {
		return err
	}
	es.decrypted = dec
	return nil
}

func (es *EncryptedString) Get() string {
	if es == nil {
		return ""
	}
	return es.decrypted
}

func (es *EncryptedString) Set(to string) {
	es.encrypted, es.decrypted = "", to
}
