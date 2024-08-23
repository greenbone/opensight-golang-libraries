// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MyTable struct {
	gorm.Model
	Field1   string
	PwdField string `encrypt:"true"`
}

var cryptor = DBCrypt[MyTable]{}

func (a *MyTable) encrypt(tx *gorm.DB) (err error) {
	err = cryptor.EncryptStruct(a)
	if err != nil {
		err := tx.AddError(fmt.Errorf("unable to encrypt password %w", err))
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func (a *MyTable) BeforeCreate(tx *gorm.DB) (err error) {
	return a.encrypt(tx)
}

func (a *MyTable) BeforeUpdate(tx *gorm.DB) (err error) {
	return a.encrypt(tx)
}

func (a *MyTable) BeforeSave(tx *gorm.DB) (err error) {
	return a.encrypt(tx)
}

func (a *MyTable) AfterFind(tx *gorm.DB) (err error) {
	err = cryptor.DecryptStruct(a)
	if err != nil {
		err := tx.AddError(fmt.Errorf("unable to decrypt password %w", err))
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func getTestDb(t *testing.T) *gorm.DB {
	// db, err := gorm.Open(sqlite.Open("file:memory:?"), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&MyTable{})
	require.NoError(t, err)
	return db
}

func TestEncryptDecrypt(t *testing.T) {
	os.Setenv("TASK_REPORT_CRYPTO_V1_PASSWORD", "my-key-1234567890")
	os.Setenv("TASK_REPORT_CRYPTO_V1_SALT", "my-salt-0987654321-0987654321-09")
	defer func() {
		os.Unsetenv("TASK_REPORT_CRYPTO_V1_PASSWORD")
		os.Unsetenv("TASK_REPORT_CRYPTO_V1_SALT")
	}()

	clearData := &MyTable{
		Field1:   "111111111",
		PwdField: "ThePassword",
	}
	originalPw := clearData.PwdField

	cryptor := DBCrypt[MyTable]{}
	err := cryptor.EncryptStruct(clearData)
	require.NoError(t, err)
	require.NotEqual(t, originalPw, clearData.PwdField, "password was not encrypted")
	err = cryptor.DecryptStruct(clearData)
	require.NoError(t, err)
	assert.Equal(t, originalPw, clearData.PwdField)
}

func TestApplianceEncryption(t *testing.T) {
	os.Setenv("TASK_REPORT_CRYPTO_V1_PASSWORD", "my-key-1234567890")
	os.Setenv("TASK_REPORT_CRYPTO_V1_SALT", "my-salt-0987654321-0987654321-09")
	defer func() {
		os.Unsetenv("TASK_REPORT_CRYPTO_V1_PASSWORD")
		os.Unsetenv("TASK_REPORT_CRYPTO_V1_SALT")
	}()

	myDB := getTestDb(t)
	tblData := &MyTable{
		Field1:   "ajdf",
		PwdField: "thePasswordWhichCanBeEncrypted",
	}
	myDB.Create(tblData)
	assert.NotNil(t, tblData.ID)

	resultData := &MyTable{}
	myDB.First(&resultData, tblData.ID)
	assert.EqualValues(t, "thePasswordWhichCanBeEncrypted", resultData.PwdField)
}
