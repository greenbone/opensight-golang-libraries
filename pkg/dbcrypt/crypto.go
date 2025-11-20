// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

// EncryptString encrypts the given string using the provided DBCryptV2.
func EncryptString(c *DBCryptV2, plaintext string) (string, error) {
	fmt.Println("--debug before encryption", plaintext) // TODO: Remove me
	ciphertext, err := c.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	fmt.Println("--debug after encryption", string(ciphertext)) // TODO: Remove me
	return string(ciphertext), nil
}

// DecryptString decrypts the given string using the provided DBCryptV2.
func DecryptString(c *DBCryptV2, ciphertextWithPrefix string) (string, error) {
	fmt.Println("--debug before decryption", ciphertextWithPrefix) // TODO: Remove me
	plaintext, err := c.Decrypt([]byte(ciphertextWithPrefix))
	if err != nil {
		return "", err
	}
	fmt.Println("--debug after decryption", string(plaintext)) // TODO: Remove me
	return string(plaintext), nil
}

func asPointerToStruct(x any) reflect.Value {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Pointer {
		return reflect.Value{}
	}
	el := v.Elem()
	if el.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	return v
}

func shouldStructFieldBeEncrypted(sf reflect.StructField) (bool, error) {
	encryptTag, has := sf.Tag.Lookup("encrypt")
	if !has || encryptTag == "false" {
		return false, nil
	}
	if encryptTag != "true" {
		return false, fmt.Errorf("invalid value for 'encrypt' field tag %q", encryptTag)
	}
	if !sf.IsExported() {
		return false, errors.New("unexported field marked for encryption")
	}
	if sf.Type.Kind() != reflect.String {
		return false, errors.New("invalid type of field marked for encryption")
	}
	return true, nil
}

func validateHistoricalTags(sf reflect.StructField) error {
	if _, has := sf.Tag.Lookup("encrypt"); has {
		return errors.New("support 'encrypt' struct filed tag has been removed with new DBCrypt package, use dbcrypt.EncryptedString type instead")
	}
	return nil
}

// EncryptStruct encrypts all fields withing the given struct that are tagged with 'encrypt:"true"' using the provided DBCryptV2.
func EncryptStruct(c *DBCryptV2, plaintext any) error {
	v := asPointerToStruct(plaintext)
	if !v.IsValid() {
		return errors.New("invalid value provided to struct encryption (expected a pointer to struct)")
	}
	v = v.Elem()
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		f, fTyp := v.Field(i), typ.Field(i)
		doEnc, err := shouldStructFieldBeEncrypted(fTyp)
		if err != nil {
			return fmt.Errorf("field %q: %w", fTyp.Name, err)
		}
		if !doEnc {
			continue
		}
		ciphertext, err := EncryptString(c, f.String())
		if err != nil {
			return fmt.Errorf("field %q: %w", fTyp.Name, err)
		}
		f.SetString(ciphertext)
	}
	return nil
}

func EncryptAny(c *DBCryptV2, plaintext any) error {
	value := reflect.ValueOf(plaintext)
	if value.Kind() == reflect.Pointer && value.Type().Elem().Kind() == reflect.Struct {
		return encryptRecursive(c, value)
	}
	if value.Kind() == reflect.Map {
		return encryptRecursive(c, value)
	}
	return errors.New("invalid value provided for encryption")
}

func encryptRecursive(c *DBCryptV2, plaintext reflect.Value) error {
	if es, ok := plaintext.Interface().(EncryptedString); ok {
		if err := es.Encrypt(c); err != nil {
			return err
		}
		plaintext.Set(reflect.ValueOf(es))
		return nil
	}
	if plaintext.Kind() == reflect.Pointer || plaintext.Kind() == reflect.Interface {
		if plaintext.IsNil() {
			return nil
		}
		return encryptRecursive(c, plaintext.Elem())
	}
	if plaintext.Kind() == reflect.Struct {
		typ := plaintext.Type()
		for i := 0; i < typ.NumField(); i++ {
			fTyp := typ.Field(i)
			if !fTyp.IsExported() {
				continue
			}
			if err := validateHistoricalTags(fTyp); err != nil {
				return fmt.Errorf("field %q: %w", fTyp.Name, err)
			}
			if err := encryptRecursive(c, plaintext.Field(i)); err != nil {
				return fmt.Errorf("field %q: %w", fTyp.Name, err)
			}
		}
	}
	if plaintext.Kind() == reflect.Map {
		for k, v := range plaintext.Seq2() {
			if err := encryptRecursive(c, v); err != nil {
				return fmt.Errorf("map key %q: %w", k.String(), err)
			}
		}
	}
	return nil
}

func DecryptAny(c *DBCryptV2, ciphertext any) error {
	value := reflect.ValueOf(ciphertext)
	if value.Kind() == reflect.Pointer && value.Type().Elem().Kind() == reflect.Struct {
		return decryptRecursive(c, value)
	}
	if value.Kind() == reflect.Map {
		return decryptRecursive(c, value)
	}
	return errors.New("invalid value provided for decryption")
}

func decryptRecursive(c *DBCryptV2, ciphertext reflect.Value) error {
	if es, ok := ciphertext.Interface().(EncryptedString); ok {
		if err := es.decrypt(c); err != nil {
			return err
		}
		ciphertext.Set(reflect.ValueOf(es))
		return nil
	}
	if ciphertext.Kind() == reflect.Pointer || ciphertext.Kind() == reflect.Interface {
		if ciphertext.IsNil() {
			return nil
		}
		return decryptRecursive(c, ciphertext.Elem())
	}
	if ciphertext.Kind() == reflect.Struct {
		typ := ciphertext.Type()
		for i := 0; i < typ.NumField(); i++ {
			fTyp := typ.Field(i)
			if !fTyp.IsExported() {
				continue
			}
			if err := validateHistoricalTags(fTyp); err != nil {
				return fmt.Errorf("field %q: %w", fTyp.Name, err)
			}
			if err := decryptRecursive(c, ciphertext.Field(i)); err != nil {
				return fmt.Errorf("field %q: %w", fTyp.Name, err)
			}
		}
	}
	if ciphertext.Kind() == reflect.Map {
		for k, v := range ciphertext.Seq2() {
			if err := decryptRecursive(c, v); err != nil {
				return fmt.Errorf("map key %q: %w", k.String(), err)
			}
		}
	}
	return nil
}

// DecryptStruct decrypts all fields withing the given struct that are tagged with 'encrypt:"true"' using the provided DBCryptV2.
func DecryptStruct(c *DBCryptV2, ciphertext any) error {
	v := asPointerToStruct(ciphertext)
	if !v.IsValid() {
		return errors.New("invalid value provided to struct decryption (expected a pointer to struct)")
	}
	v = v.Elem()
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		f, fTyp := v.Field(i), typ.Field(i)
		doEnc, err := shouldStructFieldBeEncrypted(fTyp)
		if err != nil {
			return fmt.Errorf("field %q: %w", fTyp.Name, err)
		}
		if !doEnc {
			continue
		}
		plaintext, err := DecryptString(c, f.String())
		if err != nil {
			return fmt.Errorf("field %q: %w", fTyp.Name, err)
		}
		f.SetString(plaintext)
	}
	return nil
}

func encryptDBModel(c *DBCryptV2, plaintext any) error {
	return EncryptAny(c, plaintext)
}

func decryptDBModel(c *DBCryptV2, ciphertext any) error {
	return DecryptAny(c, ciphertext)
}

// Register registers encryption and decryption callbacks for the provided data base, to perform automatically cryptographic operations on all models using EncryptStruct and DecryptStruct functions.
func Register(db *gorm.DB, c *DBCryptV2) error {
	encryptCb := func(db *gorm.DB) {
		db.AddError(encryptDBModel(c, db.Statement.Dest))
	}
	decryptCb := func(db *gorm.DB) {
		db.AddError(decryptDBModel(c, db.Statement.Dest))
	}

	if err := db.Callback().
		Create().
		Before("gorm:create").
		Register("crypto:before_create", encryptCb); err != nil {
		return err
	}
	if err := db.Callback().
		Create().
		After("gorm:create").
		Register("crypto:after_create", decryptCb); err != nil {
		return err
	}
	if err := db.Callback().
		Update().
		Before("gorm:update").
		Register("crypto:before_update", encryptCb); err != nil {
		return err
	}
	if err := db.Callback().
		Update().
		After("gorm:update").
		Register("crypto:after_update", decryptCb); err != nil {
		return err
	}
	if err := db.Callback().
		Query().
		After("gorm:query").
		Register("crypto:after_query", decryptCb); err != nil {
		return err
	}
	return nil
}

// Deregister removes any encryption and decryption callbacks for the provided data base.
func Deregister(db *gorm.DB) error {
	if err := db.Callback().
		Create().
		Before("gorm:create").
		Remove("crypto:before_create"); err != nil {
		return err
	}
	if err := db.Callback().
		Create().
		After("gorm:create").
		Remove("crypto:after_create"); err != nil {
		return err
	}
	if err := db.Callback().
		Update().
		Before("gorm:update").
		Remove("crypto:before_update"); err != nil {
		return err
	}
	if err := db.Callback().
		Update().
		After("gorm:update").
		Remove("crypto:after_update"); err != nil {
		return err
	}
	if err := db.Callback().
		Query().
		After("gorm:query").
		Remove("crypto:after_query"); err != nil {
		return err
	}
	return nil
}
