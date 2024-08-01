// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package openSearchClient provides a client for OpenSearch designed to allow easy mocking in tests.
//
// Example Usage:
//
//	clientConfig, err := config.ReadOpensearchClientConfig()
//	if err != nil {
//		return err
//	}
//
//	opensearchProjectClient, err := NewOpenSearchProjectClient(context.Background(), clientConfig)
//	if err != nil {
//		return err
//	}
//
//	client := NewClient(opensearchProjectClient, 10, 1)
//
//	query := `{"query":{"bool":{"filter":[{"term":{"oid":{"value":"1.3.6.1.4.1.25623.1.0.117842"}}}]}}}`
//	responseBody, err := client.Search(indexName, []byte(query))
//	if err != nil {
//		return err
//	}
//
//	searchResponse, err := UnmarshalSearchResponse[*Vulnerability](responseBody)
//	if err != nil {
//		return err
//	}
//
// For further usage examples see ./client_test.go.
package openSearchClient
