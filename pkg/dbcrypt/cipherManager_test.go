package dbcrypt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestCipherManager(t *testing.T) *CipherManager {
	t.Helper()

	conf := Config{
		Password:     "secret",
		PasswordSalt: "66666666666666666666666666666666",
	}
	require.NoError(t, conf.Validate())

	manager, err := NewCipherManager(conf)
	require.NoError(t, err)

	return manager
}

func TestNewCipherManager_InitializesCiphers(t *testing.T) {
	manager := newTestCipherManager(t)

	require.Len(t, manager.ciphers, 2)

	// First cipher (v1)
	require.Equal(t, "v1", manager.ciphers[0].Version)
	require.Equal(t, "ENC", manager.ciphers[0].Prefix)
	require.NotNil(t, manager.ciphers[0].Crypter)

	// Second cipher (v2)
	require.Equal(t, "v2", manager.ciphers[1].Version)
	require.Equal(t, "ENCV2", manager.ciphers[1].Prefix)
	require.NotNil(t, manager.ciphers[1].Crypter)
}

func TestCipherManager_GetLatestVersion_ReturnsLastCipherVersion(t *testing.T) {
	manager := newTestCipherManager(t)

	version := manager.GetDefaultVersion()
	require.Equal(t, "v2", version, "default version should be v2")
}

func TestCipherManager_GetByVersion(t *testing.T) {
	manager := newTestCipherManager(t)

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
			cipher, err := manager.GetByVersion(tt.version)

			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
				require.Nil(t, cipher)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cipher)
			require.Equal(t, tt.wantVersion, cipher.Version)
			require.Equal(t, tt.wantPrefix, cipher.Prefix)
			require.NotNil(t, cipher.Crypter)
		})
	}
}

func TestCipherManager_GetByPrefix(t *testing.T) {
	manager := newTestCipherManager(t)

	tests := []struct {
		name        string
		prefix      string
		wantVersion string
		wantPrefix  string
		wantErr     error
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
			wantErr: errors.New("prefix 'FOO' not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := manager.GetByPrefix(tt.prefix)

			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
				require.Nil(t, cipher)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cipher)
			require.Equal(t, tt.wantVersion, cipher.Version)
			require.Equal(t, tt.wantPrefix, cipher.Prefix)
			require.NotNil(t, cipher.Crypter)
		})
	}
}
