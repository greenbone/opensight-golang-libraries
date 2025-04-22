package openSearchClient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rs/zerolog/log"
)

type SyncUpdateClient struct {
	client           *opensearchapi.Client
	updateMaxRetries int
	updateRetryDelay time.Duration
}

func NewSyncUpdateClient(osClient *opensearchapi.Client, maxRetries int, retryDelay time.Duration) *SyncUpdateClient {
	return &SyncUpdateClient{
		client:           osClient,
		updateMaxRetries: maxRetries,
		updateRetryDelay: retryDelay,
	}
}

func (s *SyncUpdateClient) Update(indexName string, requestBody []byte) ([]byte, error) {
	log.Debug().Msgf("sync update requestBody: %s", string(requestBody))

	var updateResponse *opensearchapi.UpdateByQueryResp
	var result []byte
	var err error

	for i := 0; i < s.updateMaxRetries; i++ {
		updateResponse, err = s.client.UpdateByQuery(
			context.Background(),
			opensearchapi.UpdateByQueryReq{
				Indices: []string{indexName},
				Body:    bytes.NewReader(requestBody),
				Params: opensearchapi.UpdateByQueryParams{
					Pretty:  true,
					Refresh: opensearchapi.ToPointer(true),
				},
			},
		)
		if err != nil {
			log.Warn().Err(err).Msgf("attempt %d: error in UpdateByQuery", i+1)
			time.Sleep(s.updateRetryDelay)
			continue
		}

		body := updateResponse.Inspect().Response.Body
		result, err = io.ReadAll(body)
		if body != nil {
			body.Close()
		}
		if err != nil {
			log.Warn().Err(err).Msgf("attempt %d: error in io.ReadAll", i+1)
			time.Sleep(s.updateRetryDelay)
			continue
		}

		log.Debug().Msgf("attempt %d: sync update successful", i+1)

		var responseMap map[string]interface{}
		if err := json.Unmarshal(result, &responseMap); err != nil {
			return nil, fmt.Errorf("failed to unmarschal result: %w", err)
		}
		if failures, ok := responseMap["failures"]; ok {
			if len(failures.([]interface{})) > 0 {
				err = fmt.Errorf("sync update returned failures: %v", failures)
				log.Warn().Msgf("attempt %d: %v", i+1, err)
				time.Sleep(s.updateRetryDelay)
				continue
			}
		}

		return result, nil
	}

	return nil, err
}
