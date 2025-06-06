// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rs/zerolog/log"
)

// Client is a client for OpenSearch designed to allow easy mocking in tests.
// It is a wrapper around the official OpenSearch client github.com/opensearch-project/opensearch-go .
type Client struct {
	openSearchProjectClient *opensearchapi.Client
	updateQueue             *UpdateQueue
	syncUpdate              *SyncUpdateClient
}

// NewClient creates a new OpenSearch client.
//
// openSearchProjectClient is the official OpenSearch client to wrap. Use NewOpenSearchProjectClient to create it.
// updateMaxRetries is the number of retries for update requests.
// updateRetryDelay is the delay between retries.
func NewClient(openSearchProjectClient *opensearchapi.Client, updateMaxRetries int, updateRetryDelay time.Duration) *Client {
	c := &Client{
		openSearchProjectClient: openSearchProjectClient,
	}
	c.updateQueue = NewRequestQueue(openSearchProjectClient, updateMaxRetries, updateRetryDelay)
	c.syncUpdate = NewSyncUpdateClient(openSearchProjectClient, updateMaxRetries, updateRetryDelay)
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
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{indexName},
			Body:    bytes.NewReader(requestBody),
		},
	)
	if err != nil {
		return nil, err
	}

	// Get the raw response body to return a byte array.
	body := searchResponse.Inspect().Response.Body
	result, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	log.Trace().
		Int("status_code", searchResponse.Inspect().Response.StatusCode).
		Msgf(
			"search response - statusCode:'%d' json:'%s'",
			searchResponse.Inspect().Response.StatusCode, string(result),
		)

	return result, nil
}

func (c *Client) Count(indexName string, requestBody []byte) (count int64, err error) {
	log.Debug().Msgf("count requestBody: %s", string(requestBody))
	request := CountReq{
		Indices: []string{indexName},
		Body:    bytes.NewReader(requestBody),
	}
	countRequest, err := request.GetRequest()
	if err != nil {
		return 0, err
	}
	response, err := c.openSearchProjectClient.Client.Perform(countRequest)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(response.Body)
		log.Warn().Msgf("count response - statusCode:'%d' body:'%s'", response.StatusCode, responseBody)
	}

	var countResp CountResp
	if err := json.NewDecoder(response.Body).Decode(&countResp); err != nil {
		log.Error().Msgf("error decoding count response: %v", err)
		return 0, err
	}

	return countResp.Count, nil
}

func (c *Client) SearchStream(indexName string, requestBody []byte, scrollTimeout time.Duration, ctx context.Context) (io.Reader, error) {
	reader, writer := io.Pipe()
	startSignal := make(chan error, 1)

	log.Debug().Msgf("searchStream requestBody: %s", string(requestBody))

	go func() {
		var scrollID string
		// Initialize query with scroll
		searchResponse, err := c.openSearchProjectClient.Search(
			ctx,
			&opensearchapi.SearchReq{
				Indices: []string{indexName},
				Body:    bytes.NewReader(requestBody),
				Params: opensearchapi.SearchParams{
					Scroll: scrollTimeout,
				},
			},
		)
		if err != nil {
			writer.Close()
			startSignal <- err
			return
		}
		if searchResponse.Errors || searchResponse.Inspect().Response.IsError() {
			writer.Close()
			startSignal <- fmt.Errorf("search failed")
			log.Error().Msgf("search response: %s: %s",
				searchResponse.Inspect().Response.Status(),
				searchResponse.Inspect().Response.String())
			return
		}

		if searchResponse.ScrollID == nil {
			writer.Close()
			startSignal <- fmt.Errorf("search response contained no scroll ID")
			return
		}

		startSignal <- nil

		scrollID = *searchResponse.ScrollID
		body := searchResponse.Inspect().Response.Body
		defer body.Close()

		_, err = io.Copy(writer, body)
		if err != nil {
			writer.CloseWithError(err)
			return
		}

		// Continue scrolling thru
		scrolled := 0
		for {
			scrolled++
			log.Debug().Msgf("Scrolling %d", scrolled)
			scrollReq := opensearchapi.ScrollGetReq{
				ScrollID: scrollID,
				Params: opensearchapi.ScrollGetParams{
					Scroll: scrollTimeout,
				},
			}

			scrollResult, err := c.openSearchProjectClient.Scroll.Get(ctx, scrollReq)
			if err != nil {
				writer.CloseWithError(err)
				log.Err(err).Msgf("scroll-request failed: %v", scrollReq)
				return
			}

			if scrollResult.Inspect().Response.IsError() {
				writer.CloseWithError(fmt.Errorf("scroll-result error"))
				log.Error().Msgf("search response: %s: %s",
					scrollResult.Inspect().Response.Status(),
					scrollResult.Inspect().Response.String())
				return
			}

			noMoreHits, err := processResponse(scrollResult, writer)
			if err != nil {
				writer.CloseWithError(err)
				log.Err(err).Msgf("process response failed")
				return
			}
			if noMoreHits {
				break
			}

			// update the scrollId from the last result
			if scrollResult != nil && scrollResult.ScrollID != nil {
				scrollID = *scrollResult.ScrollID
			} else {
				log.Warn().Msg("No scroll ID found in response")
			}
		}

		writer.Close()
		// Delete Scroll Context manually
		clearScrollReq := opensearchapi.ScrollDeleteReq{
			ScrollIDs: []string{scrollID},
		}
		_, err = c.openSearchProjectClient.Scroll.Delete(context.Background(), clearScrollReq)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to delete scroll context")
		}
	}()
	err := <-startSignal
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (c *Client) CompositeAggStream(indexName string, requestBody []byte, ctx context.Context) (io.Reader, error) {
	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()
		var afterKey map[string]string

		// Loop to handle pagination using the "after" key
		loopCount := 0
		log.Trace().Msg("Starting composite aggregation stream")

		for {
			log.Trace().Msgf("Looping over compound query: %d", loopCount)
			// Build the request body dynamically by injecting the `afterKey`
			paginatedRequestBody, err := injectAfterKey(requestBody, afterKey)
			if err != nil {
				log.Err(err).Msgf("failed to inject after key: %v", requestBody)
				return
			}
			searchResponse, err := c.openSearchProjectClient.Search(
				ctx,
				&opensearchapi.SearchReq{
					Indices: []string{indexName},
					Body:    bytes.NewReader(paginatedRequestBody),
				},
			)
			if err != nil {
				log.Err(err).Msgf("search request failed: %v", requestBody)
				return
			}
			// Signal start before processing the response
			if searchResponse.Inspect().Response.IsError() {
				log.Error().Msgf("search response error: %s: %s",
					searchResponse.Inspect().Response.Status(),
					searchResponse.Inspect().Response.String())
				return
			}

			// Write the current batch of results to the writer
			body := searchResponse.Inspect().Response.Body
			responseDate, err := io.ReadAll(body)
			if err != nil {
				log.Err(err).Msgf("failed to read response body: %v", searchResponse)
				return
			}
			body.Close()
			_, err = writer.Write(responseDate)
			if err != nil {
				log.Err(err).Msgf("failed to write response to writer: %v", searchResponse)
				return
			}

			// Extract the "after" key for pagination
			afterKey, err = extractAfterKeyFromBytes(responseDate)
			if err != nil {
				log.Err(err).Msgf("failed to extract after key from response: %v", searchResponse)
				return
			}

			// If no more `afterKey`, break the loop
			if afterKey == nil {
				return
			}
			loopCount++
		}
	}()

	return reader, nil
}

func injectAfterKey(requestBody []byte, afterKey map[string]string) ([]byte, error) {
	var body map[string]interface{}
	err := json.Unmarshal(requestBody, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request body: %w", err)
	}

	aggs, ok := body["aggs"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no aggregations found in request body")
	}

	// Iterate over all aggregations to find the first composite aggregation
	for _, agg := range aggs {
		compositeAgg, ok := agg.(map[string]interface{})
		if !ok {
			continue
		}

		compositeConfig, ok := compositeAgg["composite"].(map[string]interface{})
		if ok {
			if afterKey != nil {
				compositeConfig["after"] = afterKey
			} else {
				delete(compositeConfig, "after")
			}
			return json.Marshal(body)
		}
	}
	return nil, fmt.Errorf("no composite aggregation found in request body")
}

func extractAfterKeyFromBytes(responseData []byte) (map[string]string, error) {
	var body map[string]interface{}
	err := json.Unmarshal(responseData, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response data: %w", err)
	}

	aggregations, ok := body["aggregations"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no aggregations found in response")
	}

	// Iterate over all aggregations to find the first with an "after_key"
	for _, agg := range aggregations {
		compositeAgg, ok := agg.(map[string]interface{})
		if !ok {
			continue
		}

		afterKey, ok := compositeAgg["after_key"].(map[string]interface{})
		if ok && len(afterKey) > 0 {
			// Convert `afterKey` to map[string]string
			stringAfterKey := make(map[string]string)
			for k, v := range afterKey {
				stringAfterKey[k] = fmt.Sprintf("%v", v)
			}
			return stringAfterKey, nil
		}
	}

	// No `after_key` found in any aggregation
	return nil, nil
}

// processResponse reads the response, checks for hits, and writes them to the writer
func processResponse(response *opensearchapi.ScrollGetResp, writer *io.PipeWriter) (noMoreHits bool, err error) {
	if len(response.Hits.Hits) <= 0 {
		return true, nil
	}
	body := response.Inspect().Response.Body
	_, err = io.Copy(writer, body)
	body.Close()
	if err != nil {
		return false, err
	}
	return false, nil
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

// SyncUpdate updates documents in the given index synchronously.
func (c *Client) SyncUpdate(indexName string, requestBody []byte) (responseBody []byte, err error) {
	return c.syncUpdate.Update(indexName, requestBody)
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
	waitForCompletion := !isAsync

	params := opensearchapi.DocumentDeleteByQueryParams{
		WaitForCompletion: &waitForCompletion,
	}

	if waitForCompletion {
		// Only add refresh=true when we wait for the deletion to complete
		params.Refresh = opensearchapi.ToPointer(true)
	}

	resp, err := c.openSearchProjectClient.Document.DeleteByQuery(
		context.Background(),
		opensearchapi.DocumentDeleteByQueryReq{
			Indices: []string{indexName},
			Body:    bytes.NewReader(requestBody),
			Params:  params,
		},
	)
	if err != nil {
		return err
	}
	defer resp.Inspect().Response.Body.Close()

	if resp.Inspect().Response.IsError() {
		return fmt.Errorf("error while deleting documents in index %s: %s",
			indexName, resp.Inspect().Response.String())
	}

	return nil
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
			return nil, err
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
	resp, err := c.openSearchProjectClient.Bulk(
		context.Background(),
		opensearchapi.BulkReq{
			Index: indexName,
			Body:  bytes.NewReader(requestBody),
			Params: opensearchapi.BulkParams{
				Refresh: "false",
			},
		},
	)
	if err != nil {
		return err
	}
	defer resp.Inspect().Response.Body.Close()

	if resp.Inspect().Response.IsError() {
		return fmt.Errorf("error while performing bulk update on index %s: %s",
			indexName, resp.Inspect().Response.String())
	}

	return nil
}

// Close stops the underlying UpdateQueue allowing a graceful shutdown.
func (c *Client) Close() {
	c.updateQueue.Stop()
}
