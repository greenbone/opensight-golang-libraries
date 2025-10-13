// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

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

func TestIndexCheck(t *testing.T) {
	ctx := context.Background()
	opensearchContainer, conf, err := StartOpensearchTestContainer(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, opensearchContainer)
	defer func() {
		opensearchContainer.Terminate(ctx)
	}()

	// Init OpenSearch
	client, err := NewOpenSearchProjectClient(context.Background(), conf, nil)
	require.NoError(t, err)

	// Init Index
	iFunc := NewIndexFunction(client)

	// There shall be no index at the beginning
	doesNotExists, err := iFunc.IndexExists("test")
	require.NoError(t, err)
	assert.Equal(t, false, doesNotExists)

	err = iFunc.CreateIndex("testindex", []byte(testIndex))
	assert.NoError(t, err)

	// Alias does not exist
	exists, err := iFunc.AliasExists("aliasName")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Now add the alias
	err = iFunc.CreateOrPutAlias("aliasName", "testindex")
	assert.NoError(t, err)

	// Check if it now exists
	exists, err = iFunc.AliasExists("aliasName")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Now delete the alias
	err = iFunc.DeleteAliasFromIndex("testindex", "aliasName")
	assert.NoError(t, err)

	// Now check if the alias is still there
	exists, err = iFunc.AliasExists("aliasName")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Now add a second index and add it to the alias
	err = iFunc.CreateIndex("testindex2", []byte(testIndex))
	assert.NoError(t, err)

	// Now check if we can get the settings from the index
	settings, err := iFunc.GetIndexSettings("testindex2")
	jsonData, err := json.Marshal(settings)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "settings")
	assert.Contains(t, string(jsonData), "number_of_replicas")

	// Lets write a settings back to the Index
	updatedSettings := `{"index": {"number_of_replicas": 1}}`
	fmt.Println(updatedSettings)
	err = iFunc.SetIndexSettings("testindex2", strings.NewReader(updatedSettings))
	assert.NoError(t, err)

	err = iFunc.CreateOrPutAlias("aliasName", "testindex", "testindex2")
	assert.NoError(t, err)

	// Now check if the indexes are removed from the alias
	indexes, err := iFunc.GetIndexesForAlias("aliasName")
	assert.NoError(t, err)
	assert.True(t, len(indexes) == 2)
	assert.Contains(t, indexes, "testindex")
	assert.Contains(t, indexes, "testindex2")

	// Now get the indexes by pattern
	getIndexes, err := iFunc.GetIndexes("testinde*")
	assert.NoError(t, err)
	assert.Equal(t, []string{"testindex", "testindex2"}, getIndexes)

	// Now remove the indexes from the alias
	err = iFunc.RemoveIndexesFromAlias([]string{"testindex", "testindex2"}, "aliasName")
	assert.NoError(t, err)

	// Now check if the indexes are removed from the alias
	indexes, err = iFunc.GetIndexesForAlias("aliasName")
	assert.NoError(t, err)
	assert.True(t, len(indexes) == 0)

	// Now delete the indexes
	err = iFunc.DeleteIndex("testindex")
	assert.NoError(t, err)

	err = iFunc.DeleteIndex("testindex2")
	assert.NoError(t, err)

	// Now check if the indexes are deleted
	doesNotExists, err = iFunc.IndexExists("testindex")
	require.NoError(t, err)
	assert.Equal(t, false, doesNotExists)

	doesNotExists, err = iFunc.IndexExists("testindex2")
	require.NoError(t, err)
	assert.Equal(t, false, doesNotExists)
}

func TestCreateOrPutAliasWithLargeIndexes(t *testing.T) {
	ctx := context.Background()
	currentTime := time.Now().UTC()
	opensearchContainer, conf, err := StartOpensearchTestContainer(ctx)
	require.NoError(t, err)
	require.NotNil(t, opensearchContainer)
	defer func() {
		opensearchContainer.Terminate(ctx)
	}()

	client, err := NewOpenSearchProjectClient(context.Background(), conf, nil)
	require.NoError(t, err)
	iFunc := NewIndexFunction(client)

	indexNames := make([]string, 365)
	for i := 0; i < 365; i++ {
		date := currentTime.AddDate(0, 0, -i)
		indexName := fmt.Sprintf("go_vulnerability_%s_1", date.Format("20060102"))
		indexNames[i] = indexName
		err := iFunc.CreateIndex(indexName, []byte(testIndex))
		require.NoError(t, err)
	}

	aliasName := "go_last_365_days_vulnerability"
	err = iFunc.CreateOrPutAlias(aliasName, indexNames...)
	require.NoError(t, err)

	indexes, err := iFunc.GetIndexesForAlias(aliasName)
	require.NoError(t, err)
	assert.Len(t, indexes, 365)
	for _, name := range indexNames {
		assert.Contains(t, indexes, name)
	}
}
