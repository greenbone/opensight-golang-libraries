package openSearchClient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rs/zerolog/log"
	"io"
	"time"
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
			log.Warn().Err(err).
				Int("attempt_number", i+1).
				Msgf("attempt %d: error in UpdateByQuery", i+1)
			time.Sleep(s.updateRetryDelay)
			continue
		}

		body := updateResponse.Inspect().Response.Body
		result, err = io.ReadAll(body)
		if err != nil {
			log.Warn().Err(err).
				Int("attempt_number", i+1).
				Msgf("attempt %d: error in io.ReadAll", i+1)
			time.Sleep(s.updateRetryDelay)
			continue
		}

		log.Debug().Msgf("attempt %d: sync update successful", i+1)

		var responseMap map[string]interface{}
		if err := json.Unmarshal(result, &responseMap); err != nil {
			return result, err
		}
		if failures, ok := responseMap["failures"]; ok {
			if len(failures.([]interface{})) > 0 {
				return result, fmt.Errorf("sync update failed - even after retries: %s", string(result))
			}
		}

		return result, nil
	}

	return nil, err
}
