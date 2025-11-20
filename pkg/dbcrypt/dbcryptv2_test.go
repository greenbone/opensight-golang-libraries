// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt_test

import (
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/dbcrypt"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDb[T any](t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	var table T
	err = db.AutoMigrate(&table)
	require.NoError(t, err)
	return db
}

func TestGormCreateRead(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, givenData, gotData)
}

func TestGormCreateReadNonPointerValue(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: *dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, givenData, gotData)
}

// func TestGormCreateRead(t *testing.T) {
// 	type Model struct {
// 		ID        uint   `gorm:"primarykey"`
// 		Protected string `encrypt:"true"`
// 	}
// 	db := newTestDb[Model](t)
// 	crypt, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
// 	require.NoError(t, err)
// 	require.NoError(t, dbcrypt.Register(db, crypt))

// 	givenData := Model{Protected: "aaa"}
// 	require.NoError(t, db.Create(&givenData).Error)

// 	gotData := Model{}
// 	require.NoError(t, db.First(&gotData).Error)
// 	require.Equal(t, givenData, gotData)
// }

// func TestGormCreateReadRaw(t *testing.T) {
// 	type Model struct {
// 		ID        uint   `gorm:"primarykey"`
// 		Protected string `encrypt:"true"`
// 	}
// 	db := newTestDb[Model](t)
// 	crypt, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
// 	require.NoError(t, err)
// 	require.NoError(t, dbcrypt.Register(db, crypt))

// 	givenData := Model{Protected: "aaa"}
// 	require.NoError(t, db.Create(&givenData).Error)

// 	require.NoError(t, dbcrypt.Deregister(db))

//		gotData := Model{}
//		require.NoError(t, db.First(&gotData).Error)
//		require.NotEqual(t, givenData, gotData)
//	}

func TestGormCreateReadRaw(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Protected: dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	require.NoError(t, dbcrypt.Deregister(db))

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.NotEqual(t, givenData, gotData)
	givenDataEncrypted, _ := givenData.Protected.Encrypted()
	gotDataEncrypted, _ := gotData.Protected.Encrypted()
	require.Equal(t, givenDataEncrypted, gotDataEncrypted)
}

// func TestGormCreateUpdateRead(t *testing.T) {
// 	type Model struct {
// 		ID        uint   `gorm:"primarykey"`
// 		Protected string `encrypt:"true"`
// 	}
// 	db := newTestDb[Model](t)
// 	crypt, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
// 	require.NoError(t, err)
// 	require.NoError(t, dbcrypt.Register(db, crypt))

// 	givenData := Model{Protected: "aaa"}
// 	require.NoError(t, db.Create(&givenData).Error)

// 	updatedData := Model{ID: givenData.ID, Protected: "bbb"}
// 	require.NoError(t, db.Model(&updatedData).Updates(&updatedData).Error)

// 	gotData := Model{}
// 	require.NoError(t, db.First(&gotData).Error)
// 	require.Equal(t, updatedData, gotData)
// }

func TestGormCreateUpdateRead(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
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

func TestGormCreateColumnUpdateRead(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
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

// func TestGormMixDBCryptInstances(t *testing.T) {
// 	type Model struct {
// 		ID        uint   `gorm:"primarykey"`
// 		Protected string `encrypt:"true"`
// 	}
// 	db := newTestDb[Model](t)
// 	cryptFirst, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
// 	require.NoError(t, err)
// 	cryptSecond, err := dbcrypt.New(dbcrypt.Config{Password: "other-encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
// 	require.NoError(t, err)
// 	require.NoError(t, dbcrypt.Register(db, cryptFirst))

// 	givenData := Model{Protected: "aaa"}
// 	require.NoError(t, db.Create(&givenData).Error)

// 	require.NoError(t, dbcrypt.Deregister(db))
// 	require.NoError(t, dbcrypt.Register(db, cryptSecond))
// 	require.Error(t, db.First(&Model{}).Error)
// }

func TestGormMixDBCryptInstances(t *testing.T) {
	type Model struct {
		ID        uint `gorm:"primarykey"`
		Protected *dbcrypt.EncryptedString
	}
	db := newTestDb[Model](t)
	cryptFirst, err := dbcrypt.New(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	cryptSecond, err := dbcrypt.New(dbcrypt.Config{Password: "other-encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, cryptFirst))

	givenData := Model{Protected: dbcrypt.NewEncryptedString("aaa")}
	require.NoError(t, db.Create(&givenData).Error)

	require.NoError(t, dbcrypt.Deregister(db))
	require.NoError(t, dbcrypt.Register(db, cryptSecond))
	require.Error(t, db.First(&Model{}).Error)
}
