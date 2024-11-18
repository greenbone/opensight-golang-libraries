// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"context"
	"net/http"
	"time"

	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"

	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// OpensearchTestContainer represents the opensearch container
type OpensearchTestContainer struct {
	testcontainers.Container
}

const openSearchTestDefaultHttpPort = "9200/tcp"

// StartOpensearchTestContainer starts a test container with opensearch
// and returns the container and the config for the opensearch client.
// It returns an error if the container couldn't be created or started.
//
// ctx is the context to use for the container.
func StartOpensearchTestContainer(ctx context.Context) (testcontainers.Container, config.OpensearchClientConfig, error) {
	req := testcontainers.ContainerRequest{
		Image:        "opensearchproject/opensearch:2.18.0",
		ExposedPorts: []string{openSearchTestDefaultHttpPort, "9300/tcp"},
		WaitingFor:   createWaitStrategyFor(),
		Env: map[string]string{
			"DISABLE_SECURITY_PLUGIN": "true",
			"discovery.type":          "single-node",
		},
	}
	opensearchContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Debug().Err(err).Msg("failed to create container")
	}

	host, _ := opensearchContainer.Host(ctx)
	localPort, _ := opensearchContainer.MappedPort(ctx, openSearchTestDefaultHttpPort)

	conf := config.OpensearchClientConfig{
		Host:         host,
		Port:         localPort.Int(),
		Https:        false,
		AuthMethod:   "basic",
		AuthUsername: "user",
		AuthPassword: "password",
	}

	return opensearchContainer, conf, nil
}

func createWaitStrategyFor() wait.Strategy {
	return wait.NewHTTPStrategy("/").
		WithPort(openSearchTestDefaultHttpPort).
		WithStatusCodeMatcher(func(status int) bool { return status == http.StatusOK }).
		WithStartupTimeout(10 * time.Second).
		WithStartupTimeout(5 * time.Minute)
}
