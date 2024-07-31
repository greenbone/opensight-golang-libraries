// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rs/zerolog/log"
)

type Response struct {
	Body []byte
	Err  error
}

type Request struct {
	IndexName   string
	RequestBody []byte
	Response    chan Response // Use the new Response type
}

// UpdateQueue is a queue for OpenSearch update requests.
type UpdateQueue struct {
	client           *opensearchapi.Client
	queue            chan *Request
	stop             chan bool
	wg               sync.WaitGroup
	updateMaxRetries int
	updateRetryDelay time.Duration
}

// NewRequestQueue creates a new update queue.
//
// openSearchClient is the official OpenSearch client. Use NewOpenSearchProjectClient to create it.
// updateMaxRetries is the number of retries for update requests.
// updateRetryDelay is the delay between retries.
func NewRequestQueue(openSearchClient *opensearchapi.Client, updateMaxRetries int, updateRetryDelay time.Duration) *UpdateQueue {
	rQueue := &UpdateQueue{
		client:           openSearchClient,
		queue:            make(chan *Request, 10),
		stop:             make(chan bool),
		updateMaxRetries: updateMaxRetries,
		updateRetryDelay: updateRetryDelay,
	}
	rQueue.start()
	return rQueue
}

func (q *UpdateQueue) start() {
	q.wg.Add(1)
	go q.run()
}

func (q *UpdateQueue) Stop() {
	close(q.stop)
	q.wg.Wait()
}

// Update queues and update for an index and returns the response body or an error
//
// Is called from pkg/openSearch/open_search_client/client.go:
// func (c *Client) Update(indexName string, requestBody []byte) (responseBody []byte, err error)
// and tested in pkg/openSearch/open_search_client/client_test.go
//
// indexName: The name of the index to update
// requestBody: The request body to send to the index
//
// Returns: The response body or an error
func (q *UpdateQueue) Update(indexName string, requestBody []byte) ([]byte, error) {
	request := &Request{
		IndexName:   indexName,
		RequestBody: requestBody,
		Response:    make(chan Response),
	}

	q.queue <- request

	response := <-request.Response
	close(request.Response)

	if response.Err != nil {
		return nil, response.Err
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(response.Body, &responseMap); err != nil {
		return nil, err
	}

	if _, ok := responseMap["failures"]; ok {
		if len(responseMap["failures"].([]interface{})) > 0 {
			return response.Body, fmt.Errorf("Update failed - even after retries: %s", string(response.Body))
		}
	}

	return response.Body, nil
}

func (q *UpdateQueue) run() {
	defer q.wg.Done()

	for {
		select {
		case request := <-q.queue:
			responseBody, err := q.update(request.IndexName, request.RequestBody)
			if err != nil {
				log.Error().Err(err).Msgf("Update request failed %v", responseBody)
				request.Response <- Response{Err: err}
				continue
			}
			request.Response <- Response{Body: responseBody}

		case <-q.stop:
			return
		}
	}
}

func (q *UpdateQueue) update(indexName string, requestBody []byte) ([]byte, error) {
	log.Debug().Msgf("update requestBody: %s", string(requestBody))

	var updateResponse *opensearchapi.UpdateByQueryResp
	var result []byte
	var err error

	for i := 0; i < q.updateMaxRetries; i++ {
		updateResponse, err = q.client.UpdateByQuery(
			context.Background(),
			opensearchapi.UpdateByQueryReq{
				Indices: []string{indexName},
				Body:    bytes.NewReader(requestBody),
				Params: opensearchapi.UpdateByQueryParams{
					Pretty: true,
				},
			},
		)
		if err != nil {
			log.Warn().Err(err).
				Int("attempt_number", i+1).
				Msgf("attempt %d: error in UpdateByQuery", i+1)
			time.Sleep(q.updateRetryDelay)
			continue
		}

		body := updateResponse.Inspect().Response.Body
		result, err = io.ReadAll(body)
		if err != nil {
			log.Warn().Err(err).
				Int("attempt_number", i+1).
				Msgf("attempt %d: error in io.ReadAll", i+1)
			time.Sleep(q.updateRetryDelay)
			continue
		}

		log.Debug().Msgf("attempt %d: update request successful", i+1)
		return result, nil
	}

	return nil, err
}
