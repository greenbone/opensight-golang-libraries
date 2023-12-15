// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/opensearch-project/opensearch-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Client is a client for OpenSearch designed to allow easy mocking in tests.
// It is a wrapper around the official OpenSearch client github.com/opensearch-project/opensearch-go .
type Client struct {
	openSearchProjectClient *opensearch.Client
	updateQueue             *UpdateQueue
}

// NewClient creates a new OpenSearch client.
//
// openSearchProjectClient is the official OpenSearch client to wrap. Use NewOpenSearchProjectClient to create it.
// updateMaxRetries is the number of retries for update requests.
// updateRetryDelay is the delay between retries.
func NewClient(openSearchProjectClient *opensearch.Client, updateMaxRetries int, updateRetryDelay time.Duration) *Client {
	c := &Client{
		openSearchProjectClient: openSearchProjectClient,
	}
	c.updateQueue = NewRequestQueue(openSearchProjectClient, updateMaxRetries, updateRetryDelay)
	return c
}

// Search searches for documents in the given index.
//
// indexName is the name of the index to search in.
// requestBody is the request body to send to OpenSearch.
// It returns the response body as or an error in case something went wrong.
func (c *Client) Search(indexName string, requestBody []byte) (responseBody []byte, err error) {
	log.Debug().Msgf("search requestBody: %s", string(requestBody))
	searchResponse, err := c.openSearchProjectClient.Search(
		c.openSearchProjectClient.Search.WithIndex(indexName),
		c.openSearchProjectClient.Search.WithBody(bytes.NewReader(requestBody)),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result, err := io.ReadAll(searchResponse.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	log.Trace().Msgf("search response - statusCode:'%d' json:'%s'", searchResponse.StatusCode, result)

	err = GetResponseError(searchResponse.StatusCode, result, indexName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// Update updates documents in the given index using UpdateQueue (which is also part of this package).
// It does not wait for the update to finish before returning.
// It returns the response body as or an error in case something went wrong.
//
// indexName is the name of the index to update.
// requestBody is the request body to send to OpenSearch specifying the update.
func (c *Client) Update(indexName string, requestBody []byte) (responseBody []byte, err error) {
	return c.updateQueue.Update(indexName, requestBody)
}

// AsyncDeleteByQuery updates documents in the given index asynchronously.
// It does not wait for the update to finish before returning.
// It returns an error in case something went wrong.
//
// indexName is the name of the index to delete from.
// requestBody is the request body to send to OpenSearch to identify the documents to be deleted.
func (c *Client) AsyncDeleteByQuery(indexName string, requestBody []byte) error {
	return c.deleteByQuery(indexName, requestBody, true)
}

// DeleteByQuery updates documents in the given index.
// It waits for the update to finish before returning.
// It returns an error in case something went wrong.
//
// indexName is the name of the index to delete from.
// requestBody is the request body to send to OpenSearch to identify the documents to be deleted.
func (c *Client) DeleteByQuery(indexName string, requestBody []byte) error {
	return c.deleteByQuery(indexName, requestBody, false)
}

func (c *Client) deleteByQuery(indexName string, requestBody []byte, isAsync bool) error {
	deleteResponse, err := c.openSearchProjectClient.DeleteByQuery(
		[]string{indexName},
		bytes.NewReader(requestBody),
		c.openSearchProjectClient.DeleteByQuery.WithWaitForCompletion(!isAsync),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(deleteResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(deleteResponse.StatusCode, resultString, indexName)
}

// SerializeDocumentsForBulkUpdate serializes documents for bulk update. Can be used in conjunction with BulkUpdate.
// It returns the serialized documents or an error in case something went wrong.
//
// indexName is the name of the index to update.
// documents are the documents to update.
func SerializeDocumentsForBulkUpdate[T any](indexName string, documents []T) ([]byte, error) {
	if len(documents) == 0 {
		return nil, fmt.Errorf("no documents to serialize")
	}

	var body strings.Builder
	body.Reset()

	for _, document := range documents {
		body.WriteString(fmt.Sprintf(`{"index": { "_index" : "%s"}}`,
			indexName) + "\n")
		documentJson, err := jsoniter.Marshal(document)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		body.WriteString(string(documentJson) + "\n")
	}
	return []byte(body.String()), nil
}

// BulkUpdate performs a bulk update in the given index.
// It returns an error in case something went wrong.
//
// indexName is the name of the index to update.
// requestBody is the request body to send to OpenSearch specifying the bulk update.
func (c *Client) BulkUpdate(indexName string, requestBody []byte) error {
	insertResponse, err := c.openSearchProjectClient.Bulk(
		bytes.NewReader(requestBody),
		c.openSearchProjectClient.Bulk.WithIndex(indexName),
		c.openSearchProjectClient.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(insertResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(insertResponse.StatusCode, resultString, indexName)
}

// GetResponseError checks if a response from OpenSearch indicated success and returns an error if not.
func GetResponseError(statusCode int, responseString []byte, indexName string) error {
	if statusCode >= 200 && statusCode < 300 {
		errorResponse := &BulkResponse{}
		err := jsoniter.Unmarshal(responseString, errorResponse)
		if err != nil {
			return errors.WithStack(err)
		}

		if errorResponse.HasError {
			return errors.Errorf("request error %v", errorResponse)
		}

		return nil
	}

	if statusCode == http.StatusBadRequest {
		openSearchErrorResponse := &OpenSearchErrorResponse{}
		err := jsoniter.Unmarshal(responseString, openSearchErrorResponse)
		if err != nil {
			return errors.WithStack(err)
		}

		if openSearchErrorResponse.Error.Type == "resource_already_exists_exception" {
			return NewOpenSearchResourceAlreadyExistsWithStack(
				fmt.Sprintf("Resource '%s' already exists", indexName))
		} else {
			return NewOpenSearchErrorWithStack(openSearchErrorResponse.Error.Reason)
		}
	} else {
		return NewOpenSearchErrorWithStack(string(responseString))
	}
}

// Close stops the underlying UpdateQueue allowing a graceful shutdown.
func (c *Client) Close() {
	c.updateQueue.Stop()
}
