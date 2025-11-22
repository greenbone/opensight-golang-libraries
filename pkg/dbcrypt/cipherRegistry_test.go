package dbcrypt_test

import (
	"errors"
	"testing"

	"github.com/greenbone/opensight-golang-libraries/pkg/dbcrypt"
	"github.com/stretchr/testify/require"
)

func newCipherRegistry(t *testing.T) *dbcrypt.CipherRegistry {
	t.Helper()

	conf := dbcrypt.Config{
		Password:     "secret",
		PasswordSalt: "66666666666666666666666666666666",
	}
	require.NoError(t, conf.Validate())

	registry, err := dbcrypt.NewCipherRegistry(conf)
	require.NoError(t, err)

	return registry
}

func TestNewCipherRegistry_InitializesCiphers(t *testing.T) {
	registry := newCipherRegistry(t)

	require.Len(t, registry.CipherSpecs, 2)

	// First cipher (v1)
	require.Equal(t, "v1", registry.CipherSpecs[0].Version)
	require.Equal(t, "ENC", registry.CipherSpecs[0].Prefix)
	require.NotNil(t, registry.CipherSpecs[0].Encrypter)

	// Second cipher (v2)
	require.Equal(t, "v2", registry.CipherSpecs[1].Version)
	require.Equal(t, "ENCV2", registry.CipherSpecs[1].Prefix)
	require.NotNil(t, registry.CipherSpecs[1].Encrypter)
}

func TestCipherRegistry_DefaultVersion(t *testing.T) {
	registry := newCipherRegistry(t)

	version := registry.DefaultVersion
	require.Equal(t, "v2", version, "default version should be v2")
}

func TestCipherRegistry_GetByVersion(t *testing.T) {
	registry := newCipherRegistry(t)

	tests := []struct {
		name        string
		version     string
		wantVersion string
		wantPrefix  string
		wantErr     error
	}{
		{
			name:        "existing v1",
			version:     "v1",
			wantVersion: "v1",
			wantPrefix:  "ENC",
		},
		{
			name:        "existing v2",
			version:     "v2",
			wantVersion: "v2",
			wantPrefix:  "ENCV2",
		},
		{
			name:    "unknown version",
			version: "v3",
			wantErr: errors.New("version 'v3' not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := registry.GetByVersion(tt.version)

			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
				require.Nil(t, cipher)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cipher)
			require.Equal(t, tt.wantVersion, cipher.Version)
			require.Equal(t, tt.wantPrefix, cipher.Prefix)
			require.NotNil(t, cipher.Encrypter)
		})
	}
}

func TestRegistryConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     dbcrypt.RegistryConfig
		wantErr string
	}{
		{
			name: "valid config with two CipherSpecs",
			cfg: dbcrypt.RegistryConfig{
				DefaultVersion: "v2",
				CipherSpecs: []dbcrypt.CipherSpec{
					{Version: "v1", Prefix: "ENC"},
					{Version: "v2", Prefix: "ENCV2"},
				},
			},
			wantErr: "",
		},
		{
			name: "missing default version",
			cfg: dbcrypt.RegistryConfig{
				DefaultVersion: "",
				CipherSpecs: []dbcrypt.CipherSpec{
					{Version: "v1", Prefix: "ENC"},
				},
			},
			wantErr: "default version is missing",
		},
		{
			name: "cipher without version",
			cfg: dbcrypt.RegistryConfig{
				DefaultVersion: "v1",
				CipherSpecs: []dbcrypt.CipherSpec{
					{Version: "", Prefix: "ENC"},
				},
			},
			wantErr: "cipher spec version is missing",
		},
		{
			name: "cipher without prefix",
			cfg: dbcrypt.RegistryConfig{
				DefaultVersion: "v1",
				CipherSpecs: []dbcrypt.CipherSpec{
					{Version: "v1", Prefix: ""},
				},
			},
			wantErr: "cipher spec prefix is missing",
		},
		{
			name: "duplicate cipher version",
			cfg: dbcrypt.RegistryConfig{
				DefaultVersion: "v1",
				CipherSpecs: []dbcrypt.CipherSpec{
					{Version: "v1", Prefix: "ENC"},
					{Version: "v1", Prefix: "ENC2"},
				},
			},
			wantErr: "duplicate cipher spec version 'v1'",
		},
		{
			name: "duplicate cipher prefix",
			cfg: dbcrypt.RegistryConfig{
				DefaultVersion: "v1",
				CipherSpecs: []dbcrypt.CipherSpec{
					{Version: "v1", Prefix: "ENC"},
					{Version: "v2", Prefix: "ENC"},
				},
			},
			wantErr: "duplicate cipher spec prefix 'ENC'",
		},
		{
			name: "default version not in CipherSpecs",
			cfg: dbcrypt.RegistryConfig{
				DefaultVersion: "v3",
				CipherSpecs: []dbcrypt.CipherSpec{
					{Version: "v1", Prefix: "ENC"},
					{Version: "v2", Prefix: "ENCV2"},
				},
			},
			wantErr: "default version 'v3' not found in cipher spec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCipherRegistry_GetByPrefix(t *testing.T) {
	registry := newCipherRegistry(t)

	tests := []struct {
		name        string
		prefix      string
		wantVersion string
		wantPrefix  string
		wantErr     string
	}{
		{
			name:        "existing ENC",
			prefix:      "ENC",
			wantVersion: "v1",
			wantPrefix:  "ENC",
		},
		{
			name:        "existing ENCV2",
			prefix:      "ENCV2",
			wantVersion: "v2",
			wantPrefix:  "ENCV2",
		},
		{
			name:    "unknown prefix",
			prefix:  "FOO",
			wantErr: "prefix 'FOO' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := registry.GetByPrefix(tt.prefix)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, cipher)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cipher)
			require.Equal(t, tt.wantVersion, cipher.Version)
			require.Equal(t, tt.wantPrefix, cipher.Prefix)
			require.NotNil(t, cipher.Encrypter)
		})
	}
}
