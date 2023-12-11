package open_search_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type indexFunction struct {
	client *opensearch.Client
}

func NewIndexFunction(client *opensearch.Client) *indexFunction {
	return &indexFunction{client: client}
}

// CreateIndex creates an index
func (i *indexFunction) CreateIndex(indexName string, indexSchema []byte) error {
	res := opensearchapi.IndicesCreateRequest{
		Index: indexName,
		Body:  bytes.NewReader(indexSchema),
	}
	searchResponse, err := res.Do(context.Background(), i.client)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(searchResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(searchResponse.StatusCode, resultString, indexName)
}

func (i *indexFunction) GetIndexes(pattern string) ([]string, error) {
	request := opensearchapi.IndicesGetRequest{
		Index:           []string{pattern},
		ExpandWildcards: "open",
	}

	response, err := request.Do(context.Background(), i.client)
	if err != nil {
		log.Debug().Msgf("Error while checking if index exists: %s", err)
		return nil, errors.WithStack(err)
	}

	resultString, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = GetResponseError(response.StatusCode, resultString, "")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return indexNameSliceOf(resultString)
}

func indexNameSliceOf(resultString []byte) ([]string, error) {
	indexMap := make(map[string]interface{})
	err := json.Unmarshal(resultString, &indexMap)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	indexSlice := make([]string, 0, len(indexMap))
	for key := range indexMap {
		indexSlice = append(indexSlice, key)
	}

	sort.Strings(indexSlice)

	return indexSlice, nil
}

func (i *indexFunction) IndexExists(indexName string) (bool, error) {
	includeAlias := true
	request := opensearchapi.IndicesExistsRequest{
		Index:          []string{indexName},
		AllowNoIndices: &includeAlias,
	}

	response, err := request.Do(context.Background(), i.client)
	if err != nil {
		log.Debug().Msgf("Error while checking if index exists: %s", err)
		return false, errors.WithStack(err)
	}

	if response.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return true, nil
}

func (i *indexFunction) DeleteIndex(indexName string) error {
	request := opensearchapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	response, err := request.Do(context.Background(), i.client)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(response.StatusCode, resultString, indexName)
}

func (i *indexFunction) CreateOrPutAlias(aliasName string, indexNames ...string) error {
	request := opensearchapi.IndicesPutAliasRequest{
		Index: indexNames,
		Name:  aliasName,
	}

	_, err := request.Do(context.Background(), i.client)
	if err != nil {
		log.Debug().Msgf("Error while creating and putting alias: %s", err)
		return errors.WithStack(err)
	}

	return nil
}

func (i *indexFunction) DeleteAliasFromIndex(indexName string, aliasName string) error {
	request := opensearchapi.IndicesDeleteAliasRequest{
		Index: []string{indexName},
		Name:  []string{aliasName},
	}

	response, err := request.Do(context.Background(), i.client)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(response.StatusCode, resultString, indexName)
}

func (i *indexFunction) AliasExists(aliasName string) (bool, error) {
	request := opensearchapi.CatAliasesRequest{
		Name: []string{aliasName},
	}

	response, err := request.Do(context.Background(), i.client)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if response.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if response.Header.Get("Content-Length") == `0` {
		log.Debug().Msgf("Alias %s does not exist", aliasName)
		return false, nil
	}
	return true, nil
}

// previously AliasPointsToIndex
func (i *indexFunction) GetIndexesForAlias(aliasName string) ([]string, error) {
	data := make(map[string][]string)
	request := &opensearchapi.CatAliasesRequest{
		Name: []string{aliasName},
	}

	response, err := request.Do(context.Background(), i.client)
	if err != nil {
		return []string{""}, err
	}

	result, err := io.ReadAll(response.Body)
	if err != nil {
		return []string{""}, err
	}

	lines := strings.Split(string(result), "\n")

	// Extract values line by line
	for _, line := range lines {
		parts := strings.Split(line, " ")

		// Ignore empty lines
		if len(parts) < 2 {
			continue
		}
		alias := parts[0]
		index := parts[1]
		data[alias] = append(data[alias], index)
	}

	return data[aliasName], nil
}

func (i *indexFunction) RemoveIndexesFromAlias(indexesToRemove []string, aliasName string) error {
	if len(indexesToRemove) > 0 {
		actions := i.createIndexRemovalActions(indexesToRemove, aliasName)

		actionsBytes, err := json.Marshal(actions)
		if err != nil {
			log.Debug().Msgf("Error marshaling actions to remove indexes: %s", err)
			return fmt.Errorf("error marshaling actions: %w", err)
		}

		res, err := i.client.Indices.UpdateAliases(bytes.NewReader(actionsBytes))
		if err != nil {
			return fmt.Errorf("error updating alias: %w", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Debug().Msgf("Error closing response body: %s", err)
			}
		}(res.Body)

		if res.IsError() {
			log.Debug().Msgf("Error removing non-compliant indexes from alias: %s", res.String())
			return fmt.Errorf("error removing non-compliant indexes from alias: %s", res.String())
		}
	}

	log.Debug().Msg("All non-compliant indexes removed from the alias.")
	return nil
}

func (i *indexFunction) createIndexRemovalActions(indexesToRemove []string, aliasName string) map[string]interface{} {
	var actions []map[string]map[string]string
	for _, idx := range indexesToRemove {
		action := map[string]map[string]string{
			"remove": {
				"index": idx,
				"alias": aliasName,
			},
		}
		actions = append(actions, action)
	}

	wrappedActions := map[string]interface{}{
		"actions": actions,
	}
	return wrappedActions
}
