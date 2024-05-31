// Copyright (C) Greenbone Networks GmbH
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
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

// NewOpenSearchProjectClient creates a new official OpenSearch client (package github.com/opensearch-project/opensearch-go)
// for usage NewClient.
// It returns an error if the client couldn't be created or the connection couldn't be established.
//
// ctx is the context to use for the connection.
// config is the configuration for the client.
func NewOpenSearchProjectClient(ctx context.Context, config config.OpensearchClientConfig) (*opensearch.Client, error) {
	protocol := "http"
	if config.Https {
		protocol = "https"
	}

	var client *opensearch.Client
	if err := retry.Do(
		func() error {
			openSearchApiConf := opensearch.Config{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
				},
				Addresses: []string{
					fmt.Sprintf("%s://%s:%d", protocol, config.Host, config.Port),
				},
				Header: http.Header{},
			}
			c, err := opensearch.NewClient(openSearchApiConf)
			if err != nil {
				return fmt.Errorf("search client couldn't be created: %w", err)
			}

			res := opensearchapi.PingRequest{}
			if _, err := res.Do(ctx, c); err != nil {
				return fmt.Errorf("connection to search couldn't be established: %w", err)
			}

			client = c
			return nil
		},
		retry.Context(ctx),
	); err != nil {
		return nil, fmt.Errorf("max init retries reached: %w", err)
	}

	return client, nil
}
