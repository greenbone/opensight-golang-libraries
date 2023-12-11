package open_search_client

import (
	"context"
	"testing"

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
	client, err := NewOpensearchProjectClient(context.Background(), conf)
	require.NoError(t, err)

	// Init Index
	iFunc := NewIndexFunction(client)

	// There shall be no index at the beginning
	doesNotExists, _ := iFunc.IndexExists("test")
	assert.Equal(t, false, doesNotExists)

	err = iFunc.CreateIndex("testindex", []byte(testIndex))
	assert.NoError(t, err)

	//Alias does not exist
	exists, err := iFunc.AliasExists("aliasName")
	assert.NoError(t, err)
	assert.False(t, exists)

	//Now add the alias
	err = iFunc.CreateOrPutAlias("aliasName", "testindex")
	assert.NoError(t, err)

	//Check if it now exists
	exists, err = iFunc.AliasExists("aliasName")
	assert.NoError(t, err)
	assert.True(t, exists)

	//Now delete the alias
	err = iFunc.DeleteAliasFromIndex("testindex", "aliasName")
	assert.NoError(t, err)

	//Now check if the alias is still there
	exists, err = iFunc.AliasExists("aliasName")
	assert.NoError(t, err)
	assert.False(t, exists)

	//Now add a second index and add it to the alias
	err = iFunc.CreateIndex("testindex2", []byte(testIndex))
	assert.NoError(t, err)

	err = iFunc.CreateOrPutAlias("aliasName", "testindex", "testindex2")
	assert.NoError(t, err)

	//Now check if the indexes are removed from the alias
	indexes, err := iFunc.GetIndexesForAlias("aliasName")
	assert.NoError(t, err)
	assert.True(t, len(indexes) == 2)

	//Now get the indexes by pattern
	getIndexes, err := iFunc.GetIndexes("testinde*")
	assert.NoError(t, err)
	assert.Equal(t, []string{"testindex", "testindex2"}, getIndexes)

	//Now remove the indexes from the alias
	err = iFunc.RemoveIndexesFromAlias([]string{"testindex", "testindex2"}, "aliasName")
	assert.NoError(t, err)

	//Now check if the indexes are removed from the alias
	indexes, err = iFunc.GetIndexesForAlias("aliasName")
	assert.NoError(t, err)
	assert.True(t, len(indexes) == 0)

	//Now delete the indexes
	err = iFunc.DeleteIndex("testindex")
	assert.NoError(t, err)

	err = iFunc.DeleteIndex("testindex2")
	assert.NoError(t, err)

	//Now check if the indexes are deleted
	doesNotExists, _ = iFunc.IndexExists("testindex")
	assert.Equal(t, false, doesNotExists)

	doesNotExists, _ = iFunc.IndexExists("testindex2")
	assert.Equal(t, false, doesNotExists)

}
