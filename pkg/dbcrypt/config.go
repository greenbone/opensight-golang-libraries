package dbcrypt

import (
	"errors"
)

// Config encapsulates configuration for CryptoService.
type Config struct {
	// Default version of the cryptographic algorithm. Useful for testing older historical implementations. Leave empty
	// to use the most recent version.
	//
	// Supported values:
	//   "" : Use the latest available algorithm version (recommended).
	//
	// See CipherRegistry for all versions
	Version string

	// Contains the password used deriving the encryption key
	Password string

	// Contains the salt for increasing password entropy
	PasswordSalt string
}

// Validate validates the provided config.
func (conf *Config) Validate() error {
	if conf.Password == "" {
		return errors.New("db password is empty")
	}

	if conf.PasswordSalt == "" {
		return errors.New("db password salt is empty")
	}
	if len(conf.PasswordSalt) < 32 {
		return errors.New("db password salt is too short")
	}

	return nil
}
