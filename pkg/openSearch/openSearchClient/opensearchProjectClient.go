// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/avast/retry-go/v4"
	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rs/zerolog/log"
)

// NewOpenSearchProjectClient creates a new official OpenSearch client (package github.com/opensearch-project/opensearch-go)
// for usage NewClient.
// It returns an error if the client couldn't be created or the connection couldn't be established.
//
// ctx is the context to use for the connection.
// config is the configuration for the client.
func NewOpenSearchProjectClient(ctx context.Context, config config.OpensearchClientConfig,
	tokenReceiver TokenReceiver,
) (*opensearchapi.Client, error) {
	protocol := "http"
	if config.Https {
		protocol = "https"
	}

	client, err := opensearchapi.NewClient(
		opensearchapi.Config{
			Client: opensearch.Config{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, // nolint:gosec
					},
				},
				Addresses: []string{
					fmt.Sprintf("%s://%s:%d", protocol, config.Host, config.Port),
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("search client couldn't be created: %w", err)
	}

	err = InjectAuthenticationIntoClient(client, config, tokenReceiver)
	if err != nil {
		return nil, fmt.Errorf("error injecting authentication into OpenSearch client: %w", err)
	}

	if err = retry.Do(
		func() error {
			if _, err := client.Ping(ctx, &opensearchapi.PingReq{}); err != nil {
				log.Debug().Msgf("connection to search couldn't be established: %v", err)
				return err
			}
			log.Debug().Msg("connection to search established")
			return nil
		},
		retry.Context(ctx),
	); err != nil {
		return nil, fmt.Errorf("connection to search couldn't be established: %w", err)
	}

	return client, nil
}
