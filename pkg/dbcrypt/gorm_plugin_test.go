// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"testing"

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
		ID     uint   `gorm:"primarykey"`
		Secret string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	crypt, err := NewCryptoManager(Config{
		Password:     "encryption-password",
		PasswordSalt: "encryption-password-salt-0123456",
	})
	require.NoError(t, err)
	require.NoError(t, Register(db, crypt))

	givenData := Model{
		Secret: "aaa",
	}
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
	crypt, err := NewCryptoManager(Config{
		Password:     "encryption-password",
		PasswordSalt: "encryption-password-salt-0123456",
	})
	require.NoError(t, err)
	require.NoError(t, Register(db, crypt))

	givenData := Model{Secret: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	gotData := Model{}
	require.NoError(t, db.Raw(`SELECT * FROM models LIMIT 1`).Scan(&gotData).Error)
	require.Equal(t, givenData.ID, gotData.ID)
	require.NotEqual(t, givenData.Secret, gotData.Secret)
}

func TestGormCreateUpdateRead(t *testing.T) {
	type Model struct {
		ID     uint   `gorm:"primarykey"`
		Secret string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	crypt, err := NewCryptoManager(Config{
		Password:     "encryption-password",
		PasswordSalt: "encryption-password-salt-0123456",
	})
	require.NoError(t, err)
	require.NoError(t, Register(db, crypt))

	givenData := Model{Secret: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	updatedData := Model{ID: givenData.ID, Secret: "bbb"}
	require.NoError(t, db.Updates(&updatedData).Error)

	gotData := Model{}
	require.NoError(t, db.First(&gotData).Error)
	require.Equal(t, updatedData, gotData)
}

func TestGormInvalidValueForEncrypt(t *testing.T) {
	type Model struct {
		ID     uint   `gorm:"primarykey"`
		Secret string `encrypt:"TRUE"`
	}
	db := newTestDb[Model](t)
	cryptFirst, err := NewCryptoManager(Config{
		Password:     "encryption-password",
		PasswordSalt: "encryption-password-salt-0123456",
	})
	require.NoError(t, err)
	require.NoError(t, Register(db, cryptFirst))

	givenData := Model{Secret: "aaa"}
	require.ErrorContains(t, db.Create(&givenData).Error, `field "Secret": invalid value for 'encrypt' field tag "TRUE"`)
}

func TestGormInvalidTypeForField(t *testing.T) {
	type Model struct {
		ID     uint `gorm:"primarykey"`
		Secret int  `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	cryptFirst, err := NewCryptoManager(Config{
		Password:     "encryption-password",
		PasswordSalt: "encryption-password-salt-0123456",
	})
	require.NoError(t, err)
	require.NoError(t, Register(db, cryptFirst))

	givenData := Model{
		Secret: 8,
	}
	require.ErrorContains(t, db.Create(&givenData).Error, `invalid type of field marked for encryption`)
}

func TestGormMixDBCryptInstances(t *testing.T) {
	type Model struct {
		ID     uint   `gorm:"primarykey"`
		Secret string `encrypt:"true"`
	}
	db := newTestDb[Model](t)
	cryptFirst, err := NewCryptoManager(Config{
		Password:     "encryption-password",
		PasswordSalt: "encryption-password-salt-0123456",
	})
	require.NoError(t, err)
	cryptSecond, err := NewCryptoManager(Config{
		Password:     "other-encryption-password",
		PasswordSalt: "encryption-password-salt-0123456"},
	)
	require.NoError(t, err)
	require.NoError(t, Register(db, cryptFirst))

	givenData := Model{Secret: "aaa"}
	require.NoError(t, db.Create(&givenData).Error)

	require.NoError(t, Register(db, cryptSecond))
	require.Error(t, db.First(&Model{}).Error)
}
