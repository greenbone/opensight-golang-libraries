// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package osquery

import (
	"os"
	"testing"

	"github.com/rs/zerolog/log"
)

const (
	// RunAllGoEnv when set triggers execution of all go tests
	RunAllGoEnv = "TEST_ALL_GO"
	// RunOpenSearchEnv when set triggers OpenSearch tests execution (pkg/opensearch/)
	RunOpenSearchEnv = "TEST_OPENSEARCH"
)

func TestMain(m *testing.M) {
	if os.Getenv(RunAllGoEnv) != "" || os.Getenv(RunOpenSearchEnv) != "" {
		os.Exit(m.Run())
	}
	log.Debug().Msgf("OpenSearch tests skipped, set %s=1 env to run them", RunOpenSearchEnv)
}
