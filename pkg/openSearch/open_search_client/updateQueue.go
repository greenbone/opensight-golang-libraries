package open_search_client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/opensearch-project/opensearch-go"
	"github.com/pkg/errors"
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

type requestQueue struct {
	opensearchProjectClient *opensearch.Client
	queue                   chan *Request
	stop                    chan bool
	wg                      sync.WaitGroup
	updateMaxRetries        int
	updateRetryDelay        time.Duration
}

func NewRequestQueue(opensearchProjectClient *opensearch.Client, updateMaxRetries int, updateRetryDelay time.Duration) *requestQueue {
	rQueue := &requestQueue{
		opensearchProjectClient: opensearchProjectClient,
		queue:                   make(chan *Request, 10),
		stop:                    make(chan bool),
		updateMaxRetries:        updateMaxRetries,
		updateRetryDelay:        updateRetryDelay,
	}
	rQueue.start()
	return rQueue
}

func (q *requestQueue) start() {
	q.wg.Add(1)
	go q.run()
}

func (q *requestQueue) Stop() {
	close(q.stop)
	q.wg.Wait()
}

// Update queues and update for an index and returns the response body or an error
//
// Is called from pkg/openSearch/open_search_client/client.go:
// func (c *client) Update(indexName string, requestBody []byte) (responseBody []byte, err error)
// and tested in pkg/openSearch/open_search_client/client_test.go
//
// indexName: The name of the index to update
// requestBody: The request body to send to the index
//
// Returns: The response body or an error
func (q *requestQueue) Update(indexName string, requestBody []byte) ([]byte, error) {
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
		return nil, errors.WithStack(err)
	}

	if _, ok := responseMap["failures"]; ok {
		if len(responseMap["failures"].([]interface{})) > 0 {
			return response.Body, errors.Errorf("Update failed - even after retries: %s", string(response.Body))
		}
	}

	return response.Body, nil
}

func (q *requestQueue) run() {
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

func (q *requestQueue) update(indexName string, requestBody []byte) ([]byte, error) {
	log.Debug().Msgf("update requestBody: %s", string(requestBody))

	var updateResponse *esapi.Response
	var result []byte
	var err error

	for i := 0; i < q.updateMaxRetries; i++ {
		req := esapi.UpdateByQueryRequest{
			Index:  []string{indexName},
			Body:   bytes.NewReader(requestBody),
			Pretty: true,
		}

		updateResponse, err = req.Do(context.Background(), q.opensearchProjectClient)
		if err != nil {
			log.Info().Err(err).Msgf("Attempt %d: Error in req.Do", i+1)
			time.Sleep(q.updateRetryDelay)
			continue
		}

		result, err = io.ReadAll(updateResponse.Body)
		log.Debug().Msgf("update response - statusCode:'%d' json:'%s'", updateResponse.StatusCode, result)
		if err != nil {
			log.Info().Err(err).Msgf("Attempt %d: Error in io.ReadAll", i+1)
			time.Sleep(q.updateRetryDelay)
			continue
		}

		err = GetResponseError(updateResponse.StatusCode, result, indexName)
		if err != nil {
			log.Info().Err(err).Msgf("Attempt %d: Error in GetResponseError", i+1)
			time.Sleep(q.updateRetryDelay)
			continue
		}

		log.Debug().Msgf("Attempt %d: Update request successful", i+1)
		return result, nil
	}

	return nil, errors.WithStack(err)
}
