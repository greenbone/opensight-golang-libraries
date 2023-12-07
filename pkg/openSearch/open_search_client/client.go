// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_client

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type simpleClient struct {
	client *opensearch.Client
	queue  *requestQueue
}

func NewSimpleClient(client *opensearch.Client, updateMaxRetries int, updateRetryDelay time.Duration) *simpleClient {
	c := &simpleClient{
		client: client,
	}
	c.queue = NewRequestQueue(client, updateMaxRetries, updateRetryDelay)
	return c
}

func (c *simpleClient) Search(indexName string, requestBody []byte) (responseBody []byte, err error) {
	log.Debug().Msgf("search requestBody: %s", string(requestBody))
	searchResponse, err := c.client.Search(
		c.client.Search.WithIndex(indexName),
		c.client.Search.WithBody(bytes.NewReader(requestBody)),
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

func (c *simpleClient) Update(indexName string, requestBody []byte) (responseBody []byte, err error) {
	return c.queue.Update(indexName, requestBody)
}

func (c *simpleClient) AsyncDeleteByQuery(indexName string, requestBody []byte) error {
	return c.deleteByQuery(indexName, requestBody, true)
}

func (c *simpleClient) DeleteByQuery(indexName string, requestBody []byte) error {
	return c.deleteByQuery(indexName, requestBody, false)
}

// deleteByQuery deletes documents by a query
func (c *simpleClient) deleteByQuery(indexName string, requestBody []byte, isAsync bool) error {
	deleteResponse, err := c.client.DeleteByQuery(
		[]string{indexName},
		bytes.NewReader(requestBody),
		c.client.DeleteByQuery.WithWaitForCompletion(!isAsync),
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

func (c *simpleClient) SaveToIndex(indexName string, documents [][]byte) error {
	if len(documents) == 0 {
		return nil
	}

	var body strings.Builder
	body.Reset()

	for _, document := range documents {
		body.WriteString(fmt.Sprintf(`{"index": { "_index" : "%s"}}`,
			indexName) + "\n")
		body.WriteString(string(document) + "\n")
	}

	insertResponse, err := c.client.Bulk(
		strings.NewReader(body.String()),
		c.client.Bulk.WithIndex(indexName),
		c.client.Bulk.WithRefresh("true"),
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

func (c *simpleClient) Close() {
	c.queue.Stop()
}
