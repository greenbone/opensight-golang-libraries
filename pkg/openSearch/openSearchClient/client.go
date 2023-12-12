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

type client struct {
	openSearchProjectClient *opensearch.Client
	queue                   *requestQueue
}

func NewClient(openSearchProjectClient *opensearch.Client, updateMaxRetries int, updateRetryDelay time.Duration) *client {
	c := &client{
		openSearchProjectClient: openSearchProjectClient,
	}
	c.queue = NewRequestQueue(openSearchProjectClient, updateMaxRetries, updateRetryDelay)
	return c
}

func (c *client) Search(indexName string, requestBody []byte) (responseBody []byte, err error) {
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

func (c *client) Update(indexName string, requestBody []byte) (responseBody []byte, err error) {
	return c.queue.Update(indexName, requestBody)
}

func (c *client) AsyncDeleteByQuery(indexName string, requestBody []byte) error {
	return c.deleteByQuery(indexName, requestBody, true)
}

func (c *client) DeleteByQuery(indexName string, requestBody []byte) error {
	return c.deleteByQuery(indexName, requestBody, false)
}

// deleteByQuery deletes documents by a query
func (c *client) deleteByQuery(indexName string, requestBody []byte, isAsync bool) error {
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

func SerializeDocumentsForBulkUpdate[T Identifiable](indexName string, documents []T) ([]byte, error) {
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

func (c *client) BulkUpdate(indexName string, requestBody []byte) error {
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

func (c *client) Close() {
	c.queue.Stop()
}
