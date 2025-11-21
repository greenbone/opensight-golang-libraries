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

func TestGormCreateRead(t *testing.T) {
	type Model struct {
		ID     uint   `gorm:"primarykey"`
		Secret string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Secret: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, givenData, gotData)
}

func TestGormCreateReadRaw(t *testing.T) {
	type Model struct {
		ID     uint   `gorm:"primarykey"`
		Secret string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Secret: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.Raw(`SELECT * FROM models LIMIT 1`).Scan(&gotData).Error)
	require.NotEqual(t, givenData, gotData)
}

func TestGormCreateUpdateRead(t *testing.T) {
	type Model struct {
		ID     uint   `gorm:"primarykey"`
		Secret string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	crypt, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, crypt))

	givenData := Model{Secret: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	updatedData := Model{ID: givenData.ID, Secret: "bbb"}
	require.NoError(t, db.Updates(&updatedData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, updatedData, gotData)
}

func TestGormMixDBCryptInstances(t *testing.T) {
	type Model struct {
		ID     uint   `gorm:"primarykey"`
		Secret string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	cryptFirst, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	cryptSecond, err := dbcrypt.NewDBCipher(dbcrypt.Config{Password: "other-encryption-password", PasswordSalt: "encryption-password-salt-0123456"})
	require.NoError(t, err)
	require.NoError(t, dbcrypt.Register(db, cryptFirst))

	givenData := Model{Secret: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	require.NoError(t, dbcrypt.Register(db, cryptSecond))
	require.Error(t, db.First(&Model{}).Error)
}
