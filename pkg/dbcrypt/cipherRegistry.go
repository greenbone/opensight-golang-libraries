package dbcrypt

import "fmt"

type Encrypter interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type CipherSpec struct {
	Encrypter

	Prefix  string
	Version string
}

type RegistryConfig struct {
	DefaultVersion string
	CipherSpecs    []CipherSpec
}

func (c RegistryConfig) Validate() error {
	if c.DefaultVersion == "" {
		return fmt.Errorf("default version is missing")
	}

	seenVersions := make(map[string]bool)
	seenPrefix := make(map[string]bool)

	defaultFound := false

	for _, spec := range c.CipherSpecs {
		if spec.Version == "" {
			return fmt.Errorf("cipher spec version is missing")
		}
		if spec.Prefix == "" {
			return fmt.Errorf("cipher spec prefix is missing")
		}

		if seenVersions[spec.Version] {
			return fmt.Errorf("duplicate cipher spec version '%s'", spec.Version)
		}
		seenVersions[spec.Version] = true

		if seenPrefix[spec.Prefix] {
			return fmt.Errorf("duplicate cipher spec prefix '%s'", spec.Prefix)
		}
		seenPrefix[spec.Prefix] = true

		if spec.Version == c.DefaultVersion {
			defaultFound = true
		}
	}
	if !defaultFound {
		return fmt.Errorf("default version '%s' not found in cipher specs", c.DefaultVersion)
	}

	return nil
}

type CipherRegistry struct {
	CipherSpecs    []CipherSpec
	DefaultVersion string
}

func NewCipherRegistry(conf Config) (*CipherRegistry, error) {
	registryConfig := RegistryConfig{
		CipherSpecs: []CipherSpec{
			// /!\ this list can only be extended, otherwise decryption will break for existing data
			{
				Version:   "v1",
				Prefix:    "ENC",
				Encrypter: NewCipherAesHex(conf.Password, conf.PasswordSalt),
			},
			{
				Version:   "v2",
				Prefix:    "ENCV2",
				Encrypter: NewCipherArgon2idBase64(conf.Password, conf.PasswordSalt),
			},
		},
		DefaultVersion: "v2",
	}
	if err := registryConfig.Validate(); err != nil {
		return nil, err
	}

	registry := CipherRegistry{
		CipherSpecs:    registryConfig.CipherSpecs,
		DefaultVersion: registryConfig.DefaultVersion,
	}

	return &registry, nil
}

// GetByVersion retrieves a CipherSpec by version
func (r *CipherRegistry) GetByVersion(version string) (*CipherSpec, error) {
	for _, cipher := range r.CipherSpecs {
		if cipher.Version == version {
			return &cipher, nil
		}
	}
	return nil, fmt.Errorf("version '%s' not found", version)
}

// GetByPrefix retrieves a CipherSpec by prefix
func (r *CipherRegistry) GetByPrefix(prefix string) (*CipherSpec, error) {
	for _, cipher := range r.CipherSpecs {
		if cipher.Prefix == prefix {
			return &cipher, nil
		}
	}
	return nil, fmt.Errorf("prefix '%s' not found", prefix)
}
