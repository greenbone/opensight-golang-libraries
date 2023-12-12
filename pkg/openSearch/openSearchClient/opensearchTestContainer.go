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

// Container represents the opensearch container type used in the module
type OpensearchTestContainer struct {
	testcontainers.Container
}

const openSearchTestDefaultHttpPort = "9200/tcp"

func StartOpensearchTestContainer(ctx context.Context) (testcontainers.Container, config.OpensearchClientConfig, error) {
	req := testcontainers.ContainerRequest{
		Image:        "opensearchproject/opensearch:2.11.0",
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
		log.Debug().Msgf("failed to create container: %s", err.Error())
	}

	host, _ := opensearchContainer.Host(ctx)
	localPort, _ := opensearchContainer.MappedPort(ctx, openSearchTestDefaultHttpPort)

	conf := config.OpensearchClientConfig{
		Host:  host,
		Port:  localPort.Int(),
		Https: false,
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
