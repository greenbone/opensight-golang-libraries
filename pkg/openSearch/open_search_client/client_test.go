package open_search_client

import (
	"context"
	"github.com/greenbone/opensight-golang-libraries/pkg/testFolder"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

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

	iFunc := NewIndexFunction(opensearchProjectClient)

	schema := folder.GetContent(t, "testdata/testSchema.json")
	iFunc.CreateIndex(indexName, []byte(schema))

	client := NewClient(opensearchProjectClient, 1, 1)

	t.Run("TestBulkUpdate", func(t *testing.T) {
		err = client.BulkUpdate(indexName, []byte(bulkRequest))
		require.NoError(t, err)

		responseBody, err := client.Search(indexName, []byte(``))
		require.NoError(t, err)
		assert.JSONEq(t, bulkResponse, string(responseBody))

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			responseBody, err := client.Search(indexName, []byte(``))
			require.NoError(c, err)
			assert.JSONEq(c, bulkResponse, string(responseBody))
		}, 1*time.Second, 10*time.Second)
	})

	t.Run("TestUpdate", func(t *testing.T) {

		client.Update(indexName, []byte(aVulnerabilityJson))

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, true)
		}, 1*time.Second, 10*time.Second, "external state has not changed to 'true'; still false")
	})
}

const bulkRequest = `{"index": { "_index" : "test", "_id": "someId"}}
{"oid":"1.3.6.1.4.1.25623.1.0.117842","name":"Apache Log4j 2.0.x Multiple Vulnerabilities (Windows, Log4Shell) - Version Check"}
`

const bulkResponse = `{
  "took": 1,
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
