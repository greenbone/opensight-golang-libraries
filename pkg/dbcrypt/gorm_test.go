// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/greenbone/opensight-golang-libraries/pkg/dbcrypt"
)

func newTestDb[T any](t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	var table T
	err = db.AutoMigrate(&table)
	require.NoError(t, err)
	return db
}

func TestGormCreateReadWithTag(t *testing.T) {
	type Model struct {
		ID        uint   `gorm:"primarykey"`
		Protected string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, givenData, gotData)
}

func TestGormCreateReadWithType(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, givenData, gotData)
}

func TestGormCreateReadWithTypeNonPointerValue(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: *dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, givenData, gotData)
}

func TestGormCreateReadRawWithTag(t *testing.T) {
	type Model struct {
		ID        uint   `gorm:"primarykey"`
		Protected string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.Raw(`SELECT * FROM models LIMIT 1`).Scan(&gotData).Error)
	require.NotEqual(t, givenData, gotData)
}

func TestGormCreateReadRawWithType(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.Raw(`SELECT * FROM models LIMIT 1`).Scan(&gotData).Error)
	require.NotEqual(t, givenData, gotData)
	givenDataEncrypted, _ := givenData.Protected.Encrypted()
	gotDataEncrypted, _ := gotData.Protected.Encrypted()
	require.Equal(t, givenDataEncrypted, gotDataEncrypted)
}

func TestGormCreateUpdateReadWithType(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	updatedData := Model{ID: givenData.ID, Protected: dbcrypt.NewEncryptedString("bbb")}
	require.NoError(t, db.Updates(&updatedData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, updatedData, gotData)
}

func TestGormCreateColumnUpdateReadWithType(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	updatedData := Model{ID: givenData.ID, Protected: dbcrypt.NewEncryptedString("bbb")}
	require.NoError(t, db.Model(&Model{}).Where("id = ?", updatedData.ID).Update("protected", dbcrypt.NewEncryptedString("bbb")).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	gotData.Protected.ClearEncrypted()
	require.Equal(t, updatedData, gotData)
}

func TestGormMixDBCryptInstancesWithTag(t *testing.T) {
	type Model struct {
		ID        uint   `gorm:"primarykey"`
		Protected string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	cryptFirst, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	cryptSecond, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "other-encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, cryptFirst))

	givenData := Model{Protected: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	require.NoError(t, dbcrypt.Register(db, cryptSecond))
	require.Error(t, db.First(&Model{}).Error)
}

func TestGormMixDBCryptInstancesWithType(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	cryptFirst, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	cryptSecond, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "other-encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, cryptFirst))

	givenData := Model{Protected: dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	require.NoError(t, dbcrypt.Register(db, cryptSecond))
	require.Error(t, db.First(&Model{}).Error)
}
