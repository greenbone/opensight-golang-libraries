// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_client

import (
	"bytes"
	"io"
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

func (c *simpleClient) Close() {
	c.queue.Stop()
}
