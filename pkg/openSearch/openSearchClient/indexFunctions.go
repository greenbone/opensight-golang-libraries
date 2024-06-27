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
	"sort"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type indexFunction struct {
	openSearchProjectClient *opensearchapi.Client
}

func NewIndexFunction(openSearchProjectClient *opensearchapi.Client) *indexFunction {
	return &indexFunction{openSearchProjectClient: openSearchProjectClient}
}

// CreateIndex creates an index
func (i *indexFunction) CreateIndex(indexName string, indexSchema []byte) error {
	_, err := i.openSearchProjectClient.Indices.Create(
		context.Background(),
		opensearchapi.IndicesCreateReq{
			Index: indexName,
			Body:  bytes.NewReader(indexSchema),
		},
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (i *indexFunction) GetIndexes(pattern string) ([]string, error) {
	req, err := i.openSearchProjectClient.Indices.Get(
		context.Background(),
		opensearchapi.IndicesGetReq{
			Indices: []string{pattern},
			Params: opensearchapi.IndicesGetParams{
				ExpandWildcards: "open",
			},
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body := req.Inspect().Response.Body
	defer body.Close()

	res, err := io.ReadAll(body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return indexNameSliceOf(res)
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

	req, err := i.openSearchProjectClient.Indices.Exists(
		context.Background(),
		opensearchapi.IndicesExistsReq{
			Indices: []string{indexName},
			Params: opensearchapi.IndicesExistsParams{
				AllowNoIndices: &includeAlias,
			},
		},
	)

	fmt.Printf("%v %v", req, err)

	if err != nil {
		return false, errors.WithStack(err)
	}

	return true, nil
}

func (i *indexFunction) DeleteIndex(indexName string) error {
	_, err := i.openSearchProjectClient.Indices.Delete(
		context.Background(),
		opensearchapi.IndicesDeleteReq{
			Indices: []string{indexName},
		},
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (i *indexFunction) CreateOrPutAlias(aliasName string, indexNames ...string) error {
	req, err := i.openSearchProjectClient.Indices.Alias.Put(
		context.Background(),
		opensearchapi.AliasPutReq{
			Indices: indexNames,
			Alias:   aliasName,
		},
	)
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Printf("%v %v", req, err)

	return nil
}

func (i *indexFunction) DeleteAliasFromIndex(indexName string, aliasName string) error {
	_, err := i.openSearchProjectClient.Indices.Alias.Delete(
		context.Background(),
		opensearchapi.AliasDeleteReq{
			Indices: []string{indexName},
			Alias:   []string{aliasName},
		},
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (i *indexFunction) AliasExists(aliasName string) (bool, error) {
	req, err := i.openSearchProjectClient.Cat.Aliases(
		context.Background(),
		&opensearchapi.CatAliasesReq{
			Aliases: []string{aliasName},
		},
	)

	if err != nil {
		return false, errors.WithStack(err)
	}

	if len(req.Aliases) == 0 {
		log.Debug().Str("src", "opensearch").Msgf("alias %s does not exist", aliasName)
		return false, errors.WithStack(err)
	}

	return true, nil
}

// previously AliasPointsToIndex
func (i *indexFunction) GetIndexesForAlias(aliasName string) ([]string, error) {
	req, err := i.openSearchProjectClient.Cat.Aliases(
		context.Background(),
		&opensearchapi.CatAliasesReq{
			Aliases: []string{aliasName},
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	aliases := []string{}
	for _, alias := range req.Aliases {
		aliases = append(aliases, alias.Alias)
	}

	return aliases, nil
}

func (i *indexFunction) RemoveIndexesFromAlias(indexesToRemove []string, aliasName string) error {
	if len(indexesToRemove) > 0 {
		actions := i.createIndexRemovalActions(indexesToRemove, aliasName)

		actionsBytes, err := json.Marshal(actions)
		if err != nil {
			log.Debug().Msgf("Error marshaling actions to remove indexes: %s", err)
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
