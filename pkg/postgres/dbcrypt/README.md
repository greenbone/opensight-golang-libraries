# dbcrypt

This package provides functions for encrypting and decrypting data using the AES algorithm. It uses the GCM mode of operation for encryption, which provides authentication and integrity protection for the encrypted data.
It can be used to encrypt / decrypt sensitive data using gorm hooks (see example)

## Example Usage

Here is an example of how to use the dbcrypt package:

```go
package main

import (
	"fmt"

	"github.com/example/dbcrypt"
)

type Person struct {
	gorm.Model
	Field1   string
	PwdField string `encrypt:"true"`
}

func (a *MyTable) encrypt(tx *gorm.DB) (err error) {
	err = cryptor.EncryptStruct(a)
	if err != nil {
        return err
	}
	return nil
}

func (a *MyTable) BeforeCreate(tx *gorm.DB) (err error) {
	return a.encrypt(tx)
}

func (a *MyTable) AfterFind(tx *gorm.DB) (err error) {
	err = cryptor.DecryptStruct(a)
	if err != nil {
		err := tx.AddError(fmt.Errorf("Unable to decrypt password %v", err))
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

```

In this example, a Person struct is created and encrypted using the DBCrypt struct. The encrypted struct is then saved to the database. Finally the struct is decrypted when the gorm hook is 
activated.
