package open_search_client

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/greenbone/opensight-golang-libraries/pkg/testFolder"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const (
	indexName          = "test"
	aVulnerabilityJson = `{"oid": "1.3.6.1.4.1.25623.1.0.117842", "name": "Apache Log4j 2.0.x Multiple Vulnerabilities (Windows, Log4Shell) - Version Check"}`
)

var (
	folder         = testFolder.NewTestFolder()
	aVulnerability = Vulnerability{
		Oid:  "1.3.6.1.4.1.25623.1.0.117842",
		Name: "Apache Log4j 2.0.x Multiple Vulnerabilities (Windows, Log4Shell) - Version Check",
	}
)

type ResponseStructureForHits struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
	} `json:"hits"`
}

type Vulnerability struct {
	Id   string `json:"-"`
	Oid  string `json:"oid"`
	Name string `json:"name"`
}

func (v *Vulnerability) GetId() string {
	return v.Id
}

func (v *Vulnerability) SetId(id string) {
	v.Id = id
}

func TestIndexCheck(t *testing.T) {
	ctx := context.Background()
	opensearchContainer, conf, err := StartOpensearchTestContainer(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, opensearchContainer)
	defer func() {
		opensearchContainer.Terminate(ctx)
	}()

	opensearchProjectClient, err := NewOpensearchProjectClient(context.Background(), conf)
	require.NoError(t, err)
	require.NotNil(t, opensearchProjectClient)

	iFunc := NewIndexFunction(opensearchProjectClient)

	schema := folder.GetContent(t, "testdata/testSchema.json")
	err = iFunc.CreateIndex(indexName, []byte(schema))
	assert.NoError(t, err)

	client := NewClient(opensearchProjectClient, 1, 1)

	t.Run("TestBulkUpdate", func(t *testing.T) {
		err = client.BulkUpdate(indexName, []byte(bulkRequest))
		require.NoError(t, err)

		responseBody, err := client.Search(indexName, []byte(``))
		require.NoError(t, err)
		isEqualIgnoreTookTime(t, bulkResponse, string(responseBody))

		require.Eventually(t, func() bool {
			responseBody, err := client.Search(indexName, []byte(``))
			require.NoError(t, err)
			return isEqualIgnoreTookTimeForEventually(t, bulkResponse, string(responseBody))
		}, 10*time.Second, 500*time.Millisecond)
	})

	t.Run("TestUpdate", func(t *testing.T) {
		err = client.BulkUpdate(indexName, []byte(bulkRequest))
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			responseBody, err := client.Search(indexName, []byte(``))
			require.NoError(t, err)
			return isEqualIgnoreTookTimeForEventually(t, bulkResponse, string(responseBody))
		}, 10*time.Second, 500*time.Millisecond)

		updateResponse, err := client.Update(indexName, []byte(updateRequest))
		assert.NoError(t, err)
		fmt.Println(updateResponse)

		require.Eventually(t, func() bool {
			responseBody, err := client.Search(indexName, []byte(``))
			require.NoError(t, err)
			return isEqualIgnoreTookTimeForEventually(t, updatedResponse, string(responseBody))
		}, 10*time.Second, 500*time.Millisecond, "external state has not changed to 'true'; still false")
	})

	t.Run("TestDeleteByQuery", func(t *testing.T) {
		// Add and check if its there
		err = client.BulkUpdate(indexName, []byte(bulkRequest))
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			responseBody, err := client.Search(indexName, []byte(``))
			require.NoError(t, err)
			return isEqualIgnoreTookTimeForEventually(t, bulkResponse, string(responseBody))
		}, 10*time.Second, 500*time.Millisecond)

		// Now delete it
		deleteQuery := `{"query":{"bool":{"filter":[{"term":{"oid":{"value":"1.3.6.1.4.1.25623.1.0.117842"}}}]}}}`
		err := client.AsyncDeleteByQuery(indexName, []byte(deleteQuery))
		require.NoError(t, err)

		// And now check if the delete is sucessfull
		require.Eventually(t, func() bool {
			responseBody, err := client.Search(indexName, []byte(``))
			assert.NoError(t, err)
			return hasNumberOfHits(t, string(responseBody), 0)
		}, 10*time.Second, 500*time.Millisecond, "It did not delete")
	})
}

func hasNumberOfHits(t *testing.T, reference string, hits int) bool {
	var resp ResponseStructureForHits
	err := json.Unmarshal([]byte(reference), &resp)
	require.NoError(t, err)
	if hits == resp.Hits.Total.Value {
		return true
	}
	return false
}

func isEqualIgnoreTookTimeForEventually(t *testing.T, reference string, data string) bool {
	var dataJSON map[string]interface{}
	err := json.Unmarshal([]byte(data), &dataJSON)
	require.NoError(t, err)

	delete(dataJSON, "took")

	var referenceJSON map[string]interface{}
	err = json.Unmarshal([]byte(reference), &referenceJSON)
	require.NoError(t, err)

	if reflect.DeepEqual(dataJSON, referenceJSON) {
		return true
	} else {
		return false
	}
}

func isEqualIgnoreTookTime(t *testing.T, reference string, data string) {
	var dat1 map[string]interface{}
	err := json.Unmarshal([]byte(data), &dat1)
	require.NoError(t, err)

	delete(dat1, "took")
	newResponse, err := json.Marshal(dat1)
	require.NoError(t, err)

	require.JSONEq(t, string(newResponse), reference)
}

const updateRequest = `{
  "query": {
    "match_all": {}
  },
  "script": {
    "source": "ctx._source.name = params.newName",
    "lang": "painless",
    "params": {
      "newName": "This is the New Name"
    }
  }
}`

const bulkRequest = `{"index": { "_index" : "test", "_id": "someId"}}
{"oid":"1.3.6.1.4.1.25623.1.0.117842","name":"Apache Log4j 2.0.x Multiple Vulnerabilities (Windows, Log4Shell) - Version Check"}
`

const bulkResponse = `{
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 1,
      "relation": "eq"
    },
    "max_score": 1,
    "hits": [
      {
        "_index": "test",
        "_id": "someId",
        "_score": 1,
        "_source": {
          "oid": "1.3.6.1.4.1.25623.1.0.117842",
          "name": "Apache Log4j 2.0.x Multiple Vulnerabilities (Windows, Log4Shell) - Version Check"
        }
      }
    ]
  }
}`

const updatedResponse = `{
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 1,
      "relation": "eq"
    },
    "max_score": 1,
    "hits": [
      {
        "_index": "test",
        "_id": "someId",
        "_score": 1,
        "_source": {
          "oid": "1.3.6.1.4.1.25623.1.0.117842",
          "name": "This is the New Name"
        }
      }
    ]
  }
}`
