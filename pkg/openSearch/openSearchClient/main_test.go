// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"os"
	"testing"

	"github.com/greenbone/opensight-golang-libraries/internal/testconfig"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	if os.Getenv(testconfig.RunAllGoEnv) != "" || os.Getenv(testconfig.RunOpenSearchEnv) != "" {
		os.Exit(m.Run())
	}
	log.Debug().Msgf("OpenSearch tests skipped, set %s=1 env to run them", testconfig.RunOpenSearchEnv)
}
