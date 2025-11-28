![Greenbone Logo](https://www.greenbone.net/wp-content/uploads/gb_new-logo_horizontal_rgb_small.png)

# dbcrypt Package Documentation

This package provides functions for encrypting and decrypting fields of entities persisted with GORM using the AES algorithm. It uses the GCM mode of operation for encryption, which provides authentication and integrity protection for the encrypted data. It can be used to encrypt and decrypt sensitive data using gorm hooks.

## Example Usage

Here is an example of how to use the dbcrypt package:

```go
package main

import (
	"log"

	"github.com/greenbone/opensight-golang-libraries/pkg/dbcrypt"
)

type Person struct {
	gorm.Model
	PasswordField string `encrypt:"true"`
}

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	cipher, err := dbcrypt.NewDBCipher(dbcrypt.Config{
		Password: "password",
		PasswordSalt: "password-salt-0123456789-0123456",
	})
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	dbcrypt.Register(db, cipher)

	personWrite := &Person{PasswordField: "secret"}
	if err := db.Create(personWrite).Error; err != nil {
		log.Fatalf("Error %v", err)
	}

	personRead := &Person{}
	if err := db.First(personRead).Error; err != nil {
		log.Fatalf("Error %v", err)
	}
}
```

In this example, a Person struct is created and `PasswordField` is automatically encrypted before storing in the database using the DBCipher. Then, when the data is retrieved from the database `PasswordField` is automatically decrypted.

# License

Copyright (C) 2022-2023 [Greenbone AG][Greenbone AG]

Licensed under the [GNU General Public License v3.0 or later](../../LICENSE).
