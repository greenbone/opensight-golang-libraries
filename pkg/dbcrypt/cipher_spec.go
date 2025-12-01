// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbcrypt

import (
	"fmt"
	"strings"
)

type cipherSpec struct {
	Version string
	Prefix  string
	Cipher  dbCipher
}

func (cs *cipherSpec) Validate() error {
	if cs.Version == "" {
		return fmt.Errorf("version is missing")
	}
	if cs.Prefix == "" {
		return fmt.Errorf("prefix is missing")
	}
	if strings.Contains(cs.Prefix, prefixSeparator) {
		return fmt.Errorf("prefix cannot contain %q", prefixSeparator)
	}
	if cs.Cipher == nil {
		return fmt.Errorf("cipher is missing")
	}
	return nil
}

type ciphersSpec struct {
	DefaultVersion string
	Ciphers        []cipherSpec
}

func (cs *ciphersSpec) Validate() error {
	if cs.DefaultVersion == "" {
		return fmt.Errorf("default version is missing")
	}

	seenVersions := make(map[string]bool)
	seenPrefix := make(map[string]bool)
	defaultFound := false
	for _, spec := range cs.Ciphers {
		if err := spec.Validate(); err != nil {
			return fmt.Errorf("cipher spec: %w", err)
		}
		if seenVersions[spec.Version] {
			return fmt.Errorf("duplicate cipher spec version %q", spec.Version)
		}
		seenVersions[spec.Version] = true

		if seenPrefix[spec.Prefix] {
			return fmt.Errorf("duplicate cipher spec prefix %q", spec.Prefix)
		}
		seenPrefix[spec.Prefix] = true

		if spec.Version == cs.DefaultVersion {
			defaultFound = true
		}
	}
	if !defaultFound {
		return fmt.Errorf("default version %q not found in cipher specs", cs.DefaultVersion)
	}

	return nil
}

func newCiphersSpec(conf Config) (*ciphersSpec, error) {
	cs := &ciphersSpec{
		DefaultVersion: "v2",
		Ciphers: []cipherSpec{ // /!\ this list can only be extended, otherwise decryption will break for existing data
			{
				Version: "v1",
				Prefix:  "ENC",
				Cipher:  newDbCipherHexEncode(newDbCipherGcmAesWithoutKdf(conf.Password, conf.PasswordSalt)),
			},
			{
				Version: "v2",
				Prefix:  "ENCV2",
				Cipher:  newDbCipherBase64Encode(newDbCipherGcmAesWithArgon2idKdf(conf.Password, conf.PasswordSalt)),
			},
		},
	}
	if err := cs.Validate(); err != nil {
		return nil, err
	}
	return cs, nil
}

func (cs *ciphersSpec) GetByVersion(version string) (*cipherSpec, error) {
	for _, spec := range cs.Ciphers {
		if spec.Version == version {
			return &spec, nil
		}
	}
	return nil, fmt.Errorf("cipher version %q not found", version)
}

func (cs *ciphersSpec) GetByPrefix(prefix string) (*cipherSpec, error) {
	for _, spec := range cs.Ciphers {
		if spec.Prefix == prefix {
			return &spec, nil
		}
	}
	return nil, fmt.Errorf("cipher prefix %q not found", prefix)
}
