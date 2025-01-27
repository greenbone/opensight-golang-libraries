// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
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
	"sort"
	"strings"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rs/zerolog/log"
)

type IndexFunction struct {
	openSearchProjectClient *opensearchapi.Client
}

func NewIndexFunction(openSearchProjectClient *opensearchapi.Client) *IndexFunction {
	return &IndexFunction{openSearchProjectClient: openSearchProjectClient}
}

// CreateIndex creates an index
func (i *IndexFunction) CreateIndex(indexName string, indexSchema []byte) error {
	_, err := i.openSearchProjectClient.Indices.Create(
		context.Background(),
		opensearchapi.IndicesCreateReq{
			Index: indexName,
			Body:  bytes.NewReader(indexSchema),
		},
	)
	if err != nil {
		// If the error is due to a lack of disk space or memory, we should log it as a warning
		// see details in https://repost.aws/knowledge-center/opensearch-403-clusterblockexception
		log.Err(err).Msg("error while creating index: please check disk space and memory usage")
		return err
	}

	return nil
}

func (i *IndexFunction) GetIndexes(pattern string) ([]string, error) {
	response, err := i.openSearchProjectClient.Indices.Get(
		context.Background(),
		opensearchapi.IndicesGetReq{
			Indices: []string{pattern},
			Params: opensearchapi.IndicesGetParams{
				ExpandWildcards: "open",
			},
		},
	)
	if err != nil {
		log.Debug().Err(err).Msg("error while checking if index exists")
		return nil, err
	}

	body := response.Inspect().Response.Body
	resultString, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	return indexNameSliceOf(resultString)
}

func indexNameSliceOf(resultString []byte) ([]string, error) {
	indexMap := make(map[string]interface{})
	err := json.Unmarshal(resultString, &indexMap)
	if err != nil {
		return nil, err
	}

	indexSlice := make([]string, 0, len(indexMap))
	for key := range indexMap {
		indexSlice = append(indexSlice, key)
	}

	sort.Strings(indexSlice)

	return indexSlice, nil
}

func (i *IndexFunction) IndexExists(indexName string) (bool, error) {
	includeAlias := true

	response, err := i.openSearchProjectClient.Indices.Exists(
		context.Background(),
		opensearchapi.IndicesExistsReq{
			Indices: []string{indexName},
			Params: opensearchapi.IndicesExistsParams{
				AllowNoIndices: &includeAlias,
			},
		},
	)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			return false, nil
		}

		log.Debug().Err(err).Msg("error while checking if index exists")
		return false, err
	}

	return true, nil
}

func (i *IndexFunction) DeleteIndex(indexName string) error {
	_, err := i.openSearchProjectClient.Indices.Delete(
		context.Background(),
		opensearchapi.IndicesDeleteReq{
			Indices: []string{indexName},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (i *IndexFunction) CreateOrPutAlias(aliasName string, indexNames ...string) error {
	_, err := i.openSearchProjectClient.Indices.Alias.Put(
		context.Background(),
		opensearchapi.AliasPutReq{
			Indices: indexNames,
			Alias:   aliasName,
		},
	)
	if err != nil {
		log.Debug().Err(err).Msg("error while creating and putting alias")
		return err
	}

	return nil
}

func (i *IndexFunction) DeleteAliasFromIndex(indexName string, aliasName string) error {
	_, err := i.openSearchProjectClient.Indices.Alias.Delete(
		context.Background(),
		opensearchapi.AliasDeleteReq{
			Indices: []string{indexName},
			Alias:   []string{aliasName},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (i *IndexFunction) IndexHasAlias(indexNames []string, aliasNames []string) (bool, error) {
	response, err := i.openSearchProjectClient.Indices.Alias.Exists(
		context.Background(),
		opensearchapi.AliasExistsReq{
			Indices: indexNames,
			Alias:   aliasNames,
		},
	)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			return false, nil
		}

		log.Debug().Err(err).Msg("error while checking the index alias")
		return false, err
	}

	return true, nil
}

func (i *IndexFunction) AliasExists(aliasName string) (bool, error) {
	response, err := i.openSearchProjectClient.Cat.Aliases(
		context.Background(),
		&opensearchapi.CatAliasesReq{
			Aliases: []string{aliasName},
		},
	)
	if err != nil {
		if response != nil && response.Inspect().Response.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}

	if len(response.Aliases) == 0 {
		log.Debug().Msgf("alias %s does not exist", aliasName)
		return false, nil
	}

	return true, nil
}

// previously AliasPointsToIndex
func (i *IndexFunction) GetIndexesForAlias(aliasName string) ([]string, error) {
	data := make(map[string][]string)
	response, err := i.openSearchProjectClient.Cat.Aliases(
		context.Background(),
		&opensearchapi.CatAliasesReq{
			Aliases: []string{aliasName},
		},
	)
	if err != nil {
		return nil, err
	}

	for _, alias := range response.Aliases {
		data[alias.Alias] = append(data[alias.Alias], alias.Index)
	}

	return data[aliasName], nil
}

func (i *IndexFunction) RemoveIndexesFromAlias(indexesToRemove []string, aliasName string) error {
	if len(indexesToRemove) <= 0 {
		return nil
	}

	actions := i.createIndexRemovalActions(indexesToRemove, aliasName)

	actionsBytes, err := json.Marshal(actions)
	if err != nil {
		log.Debug().Err(err).Msg("error marshaling actions to remove indexes")
		return fmt.Errorf("error marshaling actions: %w", err)
	}

	res, err := i.openSearchProjectClient.Client.Do(
		context.Background(),
		opensearchapi.AliasesReq{
			Body: bytes.NewReader(actionsBytes),
		},
		nil,
	)
	if err != nil {
		return fmt.Errorf("error updating alias: %w", err)
	}

	if res.IsError() {
		log.Debug().Msgf("error removing non-compliant indexes from alias: %s", res.String())
		return fmt.Errorf("error removing non-compliant indexes from alias: %s", res.String())
	}

	log.Debug().Msg("all non-compliant indexes removed from the alias.")
	return nil
}

func (i *IndexFunction) createIndexRemovalActions(indexesToRemove []string, aliasName string) map[string]interface{} {
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

func (i *IndexFunction) RefreshIndex(index string) error {
	log.Debug().Msgf("Start refreshing index: %s", index)
	ctx := context.Background()
	refreshResp, err := i.openSearchProjectClient.Indices.Refresh(
		ctx,
		&opensearchapi.IndicesRefreshReq{
			Indices: []string{index},
		},
	)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Index %s refreshed with staus code: %d", index,
		refreshResp.Inspect().Response.StatusCode)
	return nil
}

func (i *IndexFunction) GetIndexSettings(index string) (map[string]interface{}, error) {
	body := strings.NewReader(``)
	urlString := "/" + index + "/_settings?include_defaults=true"
	settingsRequest, err := http.NewRequest("GET", urlString, body)
	if err != nil {
		return nil, err
	}
	settingsRequest.Header.Set("Content-Type", "application/json")
	searchResp, err := i.openSearchProjectClient.Client.Perform(settingsRequest)
	if err != nil {
		return nil, err
	}
	defer searchResp.Body.Close()

	searchRespBody, err := io.ReadAll(searchResp.Body)
	if err != nil {
		return nil, err
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(searchRespBody, &settings); err != nil {
		return nil, err
	}

	// Log and return the settings for inspection
	log.Trace().Msgf("Retrieved settings for index %s: %+v", index, settings)
	return settings, nil
}

func (i *IndexFunction) SetIndexSettings(index string, settingsBody io.Reader) error {
	ctx := context.Background()

	// Apply the settings to the index
	settingsPutResp, err := i.openSearchProjectClient.Indices.Settings.Put(
		ctx,
		opensearchapi.SettingsPutReq{
			Indices: []string{index},
			Body:    settingsBody,
		},
	)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Settings applied to index %s: %t", index, settingsPutResp.Acknowledged)
	return nil
}

func (i *IndexFunction) ForceMerge(index string, maximumNumberOfSegments int) error {
	ctx := context.Background()
	forceMergeResponse, err := i.openSearchProjectClient.Indices.Forcemerge(
		ctx,
		&opensearchapi.IndicesForcemergeReq{
			Indices: []string{index},
			Params: opensearchapi.IndicesForcemergeParams{
				MaxNumSegments: &maximumNumberOfSegments,
			},
		},
	)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Forcemerge applied to index %s: with status %+v", index, forceMergeResponse.Inspect().Response)
	return nil
}
