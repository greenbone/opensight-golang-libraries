// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package secretfiles

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	// create files containing secrets
	tempDir := t.TempDir()
	validSecretPath := tempDir + "/secretFileName"
	err := os.WriteFile(validSecretPath, []byte("  secret   \n\n\t"), 0o644)
	require.NoError(t, err)
	emptySecretPath := tempDir + "/empty"
	err = os.WriteFile(emptySecretPath, []byte("  \n"), 0o644)
	require.NoError(t, err)

	type envVar struct {
		key   string
		value string
	}
	tests := map[string]struct {
		envVar  envVar
		target  *string
		want    string
		wantErr bool
	}{
		"read secrets from file": {
			target: toPtr(""),
			envVar: envVar{
				key:   "SECRET_FILE",
				value: validSecretPath,
			},
			want:    "secret",
			wantErr: false,
		},
		"existing value is overwritten": {
			target: toPtr("existing"),
			envVar: envVar{
				key:   "SECRET_FILE",
				value: validSecretPath,
			},
			want:    "secret",
			wantErr: false,
		},
		"failure with nil target": {
			target: nil,
			envVar: envVar{
				key:   "SECRET_FILE",
				value: validSecretPath,
			},
			wantErr: true,
		},
		"failure with empty secret": {
			target: toPtr(""),
			envVar: envVar{
				key:   "SECRET_FILE",
				value: emptySecretPath,
			},
			wantErr: true,
		},
		"failure with non existing file": {
			target: toPtr(""),
			envVar: envVar{
				key:   "SECRET_FILE",
				value: "/invalid/path",
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := os.Setenv(tt.envVar.key, tt.envVar.value)
			require.NoError(t, err)

			err = ReadSecret(tt.envVar.key, tt.target)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, *tt.target)
			}
		})
	}
}

func toPtr[T any](val T) *T {
	return &val
}
