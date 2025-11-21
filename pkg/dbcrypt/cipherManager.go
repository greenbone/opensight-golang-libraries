package dbcrypt

import "fmt"

type Cipher struct {
	Crypter

	Prefix  string
	Version string
}

type CipherManager struct {
	ciphers        []Cipher
	defaultVersion string
}

func NewCipherManager(conf Config) (*CipherManager, error) {
	manager := CipherManager{
		ciphers: []Cipher{
			{
				Version: "v1",
				Prefix:  "ENC",
				Crypter: NewCipherAES(conf.Password, conf.PasswordSalt),
			},
			{
				Version: "v2",
				Prefix:  "ENCV2",
				Crypter: NewCipherArgon2id(conf.Password, conf.PasswordSalt),
			},
		},
		defaultVersion: "v2",
	}

	_, err := manager.GetByVersion(manager.defaultVersion)
	if err != nil {
		return nil, fmt.Errorf("default version '%s' does not exist", manager.defaultVersion)
	}

	return &manager, nil
}

func (m *CipherManager) GetDefaultVersion() string {
	return m.defaultVersion
}

// GetByVersion retrieves a Cipher by version
func (m *CipherManager) GetByVersion(version string) (*Cipher, error) {
	for _, cipher := range m.ciphers {
		if cipher.Version == version {
			return &cipher, nil
		}
	}
	return nil, fmt.Errorf("version '%s' not found", version)
}

// GetByPrefix retrieves a Cipher by prefix
func (m *CipherManager) GetByPrefix(prefix string) (*Cipher, error) {
	for _, cipher := range m.ciphers {
		if cipher.Prefix == prefix {
			return &cipher, nil
		}
	}
	return nil, fmt.Errorf("prefix '%s' not found", prefix)
}
