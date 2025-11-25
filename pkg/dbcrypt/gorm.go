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

func parseEncryptStructFieldTag(sf reflect.StructField) (bool, error) {
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

func modelValue(model any) reflect.Value {
	value := reflect.ValueOf(model)
	for {
		switch value.Kind() {
		case reflect.Pointer, reflect.Interface:
			if value.IsNil() {
				return value
			}
			value = value.Elem()
		default:
			return value
		}
	}
}

func encryptModel(c *DBCipher, plaintext any) error {
	value := modelValue(plaintext)
	switch value.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Map:
		return encryptRecursive(c, value)
	default:
		return nil
	}
}

func encryptRecursive(c *DBCipher, plaintext reflect.Value) error {
	switch plaintext.Kind() {
	case reflect.Pointer, reflect.Interface:
		if plaintext.IsNil() {
			return nil
		}
		return encryptRecursive(c, plaintext.Elem())
	case reflect.Struct:
		typ := plaintext.Type()
		for i := 0; i < typ.NumField(); i++ {
			fTyp := typ.Field(i)
			if !fTyp.IsExported() {
				continue
			}
			if err := encryptFieldBasedOnTag(c, fTyp, plaintext.Field(i)); err != nil {
				return fmt.Errorf("field %q: %w", fTyp.Name, err)
			}
			if err := encryptRecursive(c, plaintext.Field(i)); err != nil {
				return fmt.Errorf("field %q: %w", fTyp.Name, err)
			}
		}
	case reflect.Slice:
		for i, v := range plaintext.Seq2() {
			if err := encryptRecursive(c, v); err != nil {
				return fmt.Errorf("list item #%d: %w", i.Int(), err)
			}
		}
	case reflect.Map:
		for k, v := range plaintext.Seq2() {
			if err := encryptRecursive(c, v); err != nil {
				return fmt.Errorf("map key %q: %w", k.String(), err)
			}
		}
	}
	return nil
}

func encryptFieldBasedOnTag(c *DBCipher, sf reflect.StructField, val reflect.Value) error {
	tagValue, err := parseEncryptStructFieldTag(sf)
	if err != nil {
		return err
	}
	if !tagValue {
		return nil
	}
	ciphertext, err := c.Encrypt([]byte(val.String()))
	if err != nil {
		return err
	}
	val.SetString(string(ciphertext))
	return nil
}

func decryptModel(c *DBCipher, ciphertext any) error {
	value := modelValue(ciphertext)
	switch value.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Map:
		return decryptRecursive(c, value)
	default:
		return nil
	}
}

func decryptRecursive(c *DBCipher, ciphertext reflect.Value) error {
	switch ciphertext.Kind() {
	case reflect.Pointer, reflect.Interface:
		if ciphertext.IsNil() {
			return nil
		}
		return decryptRecursive(c, ciphertext.Elem())
	case reflect.Struct:
		typ := ciphertext.Type()
		for i := 0; i < typ.NumField(); i++ {
			fTyp := typ.Field(i)
			if !fTyp.IsExported() {
				continue
			}
			if err := decryptFieldBasedOnTag(c, fTyp, ciphertext.Field(i)); err != nil {
				return fmt.Errorf("field %q: %w", fTyp.Name, err)
			}
			if err := decryptRecursive(c, ciphertext.Field(i)); err != nil {
				return fmt.Errorf("field %q: %w", fTyp.Name, err)
			}
		}
	case reflect.Slice:
		for i, v := range ciphertext.Seq2() {
			if err := decryptRecursive(c, v); err != nil {
				return fmt.Errorf("list item #%d: %w", i.Int(), err)
			}
		}
	case reflect.Map:
		for k, v := range ciphertext.Seq2() {
			if err := decryptRecursive(c, v); err != nil {
				return fmt.Errorf("map key %q: %w", k.String(), err)
			}
		}
	}
	return nil
}

func decryptFieldBasedOnTag(c *DBCipher, sf reflect.StructField, val reflect.Value) error {
	tagValue, err := parseEncryptStructFieldTag(sf)
	if err != nil {
		return err
	}
	if !tagValue {
		return nil
	}
	plaintext, err := c.Decrypt([]byte(val.String()))
	if err != nil {
		return err
	}
	val.SetString(string(plaintext))
	return nil
}

// Register registers encryption and decryption callbacks for the provided data base, to perform automatically cryptographic operations on all models that contain a field tagged with 'encrypt:"true"'.
func Register(db *gorm.DB, c *DBCipher) error {
	encryptCb := func(db *gorm.DB) {
		_ = db.AddError(encryptModel(c, &db.Statement.Dest))
	}
	decryptCb := func(db *gorm.DB) {
		_ = db.AddError(decryptModel(c, &db.Statement.Dest))
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
