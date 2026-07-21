// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/ostesting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testIndex = `{
        "settings": {
          "analysis": {
            "normalizer": {
              "keyword_lowercase": {
                "type": "custom",
                "filter": ["lowercase"]
              }
            }
          }
        },
        "mappings": {
          "dynamic": "strict",
          "properties": {
            "aProperty": {
              "type": "long"
            },
            "aDescription": {
              "index": false,
              "type": "text"
            },
            "aDate": {
              "type": "date"
            },
            "aKeyword": {
              "type": "keyword"
            }
          }
        }
      }`

func getOpenSearchConfig(t *testing.T) (*ostesting.Tester, config.OpensearchClientConfig) {
	tester := ostesting.NewTester(t, ostesting.RunNotParallelOption) // `t.Parallel()`` explicitly set in the respective testcase
	cfg := tester.Config()

	// convert config
	u, err := url.Parse(cfg.Address)
	require.NoError(t, err)
	host, portStr, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)
	conf := config.OpensearchClientConfig{
		Host:         host,
		Port:         port,
		Https:        u.Scheme == "https",
		AuthUsername: cfg.User,
		AuthPassword: cfg.Password,
		AuthMethod:   "basic",
	}

	return tester, conf
}

func TestIndexCheck(t *testing.T) {
	t.Parallel()

	tester, conf := getOpenSearchConfig(t)
	indexName := "testindex_" + strings.ToLower(rand.Text())
	aliasName := "alias_" + indexName
	t.Cleanup(func() {
		for _, indexName := range []string{indexName, indexName + "2"} {
			tester.DeleteIndex(t, indexName)
		}
	})

	// Init OpenSearch
	client, err := NewOpenSearchProjectClient(context.Background(), conf, nil)
	require.NoError(t, err)

	// Init Index
	iFunc := NewIndexFunction(client)

	// There shall be no index at the beginning
	doesNotExists, err := iFunc.IndexExists(indexName)
	require.NoError(t, err)
	assert.Equal(t, false, doesNotExists)

	err = iFunc.CreateIndex(indexName, []byte(testIndex))
	assert.NoError(t, err)

	// Alias does not exist
	exists, err := iFunc.AliasExists(aliasName)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Now add the alias
	err = iFunc.CreateOrPutAlias(aliasName, indexName)
	assert.NoError(t, err)

	// Check if it now exists
	exists, err = iFunc.AliasExists(aliasName)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Now delete the alias
	err = iFunc.DeleteAliasFromIndex(indexName, aliasName)
	assert.NoError(t, err)

	// Now check if the alias is still there
	exists, err = iFunc.AliasExists(aliasName)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Now add a second index and add it to the alias
	err = iFunc.CreateIndex(indexName+"2", []byte(testIndex))
	assert.NoError(t, err)

	// Now check if we can get the settings from the index
	settings, err := iFunc.GetIndexSettings(indexName + "2")
	assert.NoError(t, err)
	jsonData, err := json.Marshal(settings)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "settings")
	assert.Contains(t, string(jsonData), "number_of_replicas")

	// Lets write a settings back to the Index
	updatedSettings := `{"index": {"number_of_replicas": 1}}`
	fmt.Println(updatedSettings)
	err = iFunc.SetIndexSettings(indexName+"2", strings.NewReader(updatedSettings))
	assert.NoError(t, err)

	err = iFunc.CreateOrPutAlias(aliasName, indexName, indexName+"2")
	assert.NoError(t, err)

	// Now check if the indexes are removed from the alias
	indexes, err := iFunc.GetIndexesForAlias(aliasName)
	assert.NoError(t, err)
	assert.True(t, len(indexes) == 2)
	assert.Contains(t, indexes, indexName)
	assert.Contains(t, indexes, indexName+"2")

	// Now get the indexes by pattern
	getIndexes, err := iFunc.GetIndexes(indexName[:len(indexName)-1] + "*")
	assert.NoError(t, err)
	assert.Equal(t, []string{indexName, indexName + "2"}, getIndexes)

	// Now remove the indexes from the alias
	err = iFunc.RemoveIndexesFromAlias([]string{indexName, indexName + "2"}, aliasName)
	assert.NoError(t, err)

	// Now check if the indexes are removed from the alias
	indexes, err = iFunc.GetIndexesForAlias(aliasName)
	assert.NoError(t, err)
	assert.True(t, len(indexes) == 0)

	// Now delete the indexes
	err = iFunc.DeleteIndex(indexName)
	assert.NoError(t, err)

	err = iFunc.DeleteIndex(indexName + "2")
	assert.NoError(t, err)

	// Now check if the indexes are deleted
	doesNotExists, err = iFunc.IndexExists(indexName)
	require.NoError(t, err)
	assert.Equal(t, false, doesNotExists)

	doesNotExists, err = iFunc.IndexExists(indexName + "2")
	require.NoError(t, err)
	assert.Equal(t, false, doesNotExists)
}

func TestCreateOrPutAliasWithLargeIndexes(t *testing.T) {
	t.Parallel()

	currentTime := time.Now().UTC()
	tester, conf := getOpenSearchConfig(t)

	client, err := NewOpenSearchProjectClient(context.Background(), conf, nil)
	require.NoError(t, err)
	iFunc := NewIndexFunction(client)

	indexNames := make([]string, 0, 365)
	t.Cleanup(func() {
		for _, indexName := range indexNames {
			tester.DeleteIndex(t, indexName)
		}
	})

	baseIndexName := "go_vulnerability_" + strings.ToLower(rand.Text())
	for i := range 365 {
		date := currentTime.AddDate(0, 0, -i)
		indexName := fmt.Sprintf("%s_%s_1", baseIndexName, date.Format("20060102"))
		indexNames = append(indexNames, indexName)
		err := iFunc.CreateIndex(indexName, []byte(testIndex))
		require.NoError(t, err)
	}

	aliasName := "go_last_365_days_vulnerability" + strings.ToLower(rand.Text())
	err = iFunc.CreateOrPutAlias(aliasName, indexNames...)
	require.NoError(t, err)

	indexes, err := iFunc.GetIndexesForAlias(aliasName)
	require.NoError(t, err)
	assert.Len(t, indexes, 365)
	for _, name := range indexNames {
		assert.Contains(t, indexes, name)
	}
}
