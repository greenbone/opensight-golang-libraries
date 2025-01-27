// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/greenbone/opensight-golang-libraries/pkg/testFolder"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const indexName = "test"

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

func TestClient(t *testing.T) {
	type testCase struct {
		testFunc func(t *testing.T, client *Client, iFunc *IndexFunction)
	}
	tcs := map[string]testCase{
		"TestBulkUpdate": {func(t *testing.T, client *Client, iFunc *IndexFunction) {
			// given
			bulkRequest, err := SerializeDocumentsForBulkUpdate(indexName, []*Vulnerability{&aVulnerability})
			require.NoError(t, err)

			// when
			err = client.BulkUpdate(indexName, bulkRequest)

			// then
			require.NoError(t, err)

			err = iFunc.RefreshIndex(indexName)
			require.NoError(t, err)

			require.EventuallyWithT(t, func(c *assert.CollectT) {
				searchResponse := searchAllVulnerabilities(c, client)
				assert.Equal(c, uint(1), searchResponse.Hits.Total.Value)
				assert.Equal(c, 1, len(searchResponse.GetResults()))
				assert.Equal(c, aVulnerability, *searchResponse.GetResults()[0])
			}, 10*time.Second, 500*time.Millisecond)
		}},
		"TestUpdate": {func(t *testing.T, client *Client, _ *IndexFunction) {
			// given
			createDataInIndex(t, client, []*Vulnerability{&aVulnerability}, 1)
			updateRequest := `{
			  "query": {
				"match_all": {}
			  },
			  "script": {
				"source": "ctx._source.name = params.newName",
				"lang": "painless",
				"params": {
				  "newName": "This is the new name"
				}
			  }
			}`

			// when
			updateResponse, err := client.Update(indexName, []byte(updateRequest))

			// then
			require.NoError(t, err)
			log.Debug().Msgf("updateResponse: %s", string(updateResponse))
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				searchResponse := searchAllVulnerabilities(c, client)
				assert.Equal(c, uint(1), searchResponse.Hits.Total.Value)
				assert.Equal(c, 1, len(searchResponse.GetResults()))
				assert.Equal(c, "This is the new name", searchResponse.GetResults()[0].Name)
				assert.Equal(c, aVulnerability.Oid, searchResponse.GetResults()[0].Oid)
			}, 10*time.Second, 500*time.Millisecond)
		}},
		"TestAsyncDeleteByQuery": {func(t *testing.T, client *Client, _ *IndexFunction) {
			// given
			createDataInIndex(t, client, []*Vulnerability{&aVulnerability}, 1)

			// when
			deleteQuery := `{"query":{"bool":{"filter":[{"term":{"oid":{"value":"1.3.6.1.4.1.25623.1.0.117842"}}}]}}}`
			err := client.AsyncDeleteByQuery(indexName, []byte(deleteQuery))

			// then
			require.NoError(t, err)
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				searchResponse := searchAllVulnerabilities(c, client)
				assert.Equal(c, uint(0), searchResponse.Hits.Total.Value)
				assert.Equal(c, 0, len(searchResponse.GetResults()))
			}, 10*time.Second, 500*time.Millisecond)
		}},
		"TestDeleteByQuery": {func(t *testing.T, client *Client, _ *IndexFunction) {
			// given
			createDataInIndex(t, client, []*Vulnerability{&aVulnerability}, 1)

			// when
			deleteQuery := `{"query":{"bool":{"filter":[{"term":{"oid":{"value":"1.3.6.1.4.1.25623.1.0.117842"}}}]}}}`
			err := client.DeleteByQuery(indexName, []byte(deleteQuery))

			// then
			require.NoError(t, err)
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				searchResponse := searchAllVulnerabilities(c, client)
				assert.Equal(c, uint(0), searchResponse.Hits.Total.Value)
				assert.Equal(c, 0, len(searchResponse.GetResults()))
			}, 10*time.Second, 500*time.Millisecond)
		}},
		"TestSearch": {func(t *testing.T, client *Client, _ *IndexFunction) {
			// given
			createDataInIndex(t, client, []*Vulnerability{&aVulnerability}, 1)

			// when
			query := `{"query":{"bool":{"filter":[{"term":{"oid":{"value":"1.3.6.1.4.1.25623.1.0.117842"}}}]}}}`
			responseBody, err := client.Search(indexName, []byte(query))

			// then
			require.NoError(t, err)
			searchResponse, err := UnmarshalSearchResponse[*Vulnerability](responseBody)
			require.NoError(t, err)
			assert.Equal(t, uint(1), searchResponse.Hits.Total.Value)
			assert.Equal(t, 1, len(searchResponse.GetResults()))
			assert.Equal(t, aVulnerability, *searchResponse.GetResults()[0])

			// when
			query = `{"query":{"bool":{"filter":[{"term":{"oid":{"value":"doesNotExist"}}}]}}}`
			responseBody, err = client.Search(indexName, []byte(query))

			// then
			require.NoError(t, err)
			searchResponse, err = UnmarshalSearchResponse[*Vulnerability](responseBody)
			require.NoError(t, err)
			assert.Equal(t, uint(0), searchResponse.Hits.Total.Value)
			assert.Equal(t, 0, len(searchResponse.GetResults()))
		}},
		"TestSearchStream": {func(t *testing.T, client *Client, _ *IndexFunction) {
			var searchResponse SearchResponse[*Vulnerability]

			// given
			createDataInIndex(t, client, []*Vulnerability{&aVulnerability}, 1)

			// when
			query := `{"query":{"bool":{"filter":[{"term":{"oid":{"value":"1.3.6.1.4.1.25623.1.0.117842"}}}]}}}`
			responseReader, err := client.SearchStream(indexName, []byte(query), time.Millisecond, context.Background())

			// then
			require.NoError(t, err)

			decoder := json.NewDecoder(responseReader)

			// first read
			err = decoder.Decode(&searchResponse)
			require.NoError(t, err)

			// second read
			err = decoder.Decode(&searchResponse)
			require.Equal(t, io.EOF, err)

			assert.Equal(t, uint(1), searchResponse.Hits.Total.Value)
			assert.Equal(t, 1, len(searchResponse.GetResults()))
			assert.Equal(t, aVulnerability, *searchResponse.GetResults()[0])

			// when
			query = `{"query":{"bool":{"filter":[{"term":{"oid":{"value":"doesNotExist"}}}]}}}`
			responseReader, err = client.SearchStream(indexName, []byte(query), time.Millisecond, context.Background())

			// then
			require.NoError(t, err)

			decoder = json.NewDecoder(responseReader)
			err = decoder.Decode(&searchResponse)
			require.NoError(t, err)

			assert.Empty(t, searchResponse.Hits.SearchHits)
			assert.Equal(t, uint(0), searchResponse.Hits.Total.Value)
			assert.Equal(t, 0, len(searchResponse.GetResults()))
		}},
	}

	ctx := context.Background()
	opensearchContainer, conf, err := StartOpensearchTestContainer(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, opensearchContainer)
	defer func() {
		opensearchContainer.Terminate(ctx)
	}()

	opensearchProjectClient, err := NewOpenSearchProjectClient(context.Background(), conf, nil)
	require.NoError(t, err)
	require.NotNil(t, opensearchProjectClient)

	iFunc := NewIndexFunction(opensearchProjectClient)
	client := NewClient(opensearchProjectClient, 1, 1)

	for testName, testCase := range tcs {
		t.Run(testName, func(t *testing.T) {
			err = iFunc.DeleteIndex(indexName)
			if err != nil {
				log.Trace().Msgf("Error while deleting index: %s", err)
			}
			schema := folder.GetContent(t, "testdata/testSchema.json")
			err = iFunc.CreateIndex(indexName, []byte(schema))
			require.NoError(t, err)
			testCase.testFunc(t, client, iFunc)
		})
	}
}

func TestSerializeDocumentsForBulkUpdate(t *testing.T) {
	// when
	bulkUpdate, err := SerializeDocumentsForBulkUpdate(indexName, []*Vulnerability{&aVulnerability})

	// then
	require.NoError(t, err)
	expectedString := `{"index": { "_index" : "test"}}
{"oid":"1.3.6.1.4.1.25623.1.0.117842","name":"Apache Log4j 2.0.x Multiple Vulnerabilities (Windows, Log4Shell) - Version Check"}
`
	assert.Equal(t, expectedString, string(bulkUpdate))
}

func createDataInIndex(t *testing.T, client *Client, vulnerabilities []*Vulnerability, expectedDocumentCount uint) {
	bulkRequest, err := SerializeDocumentsForBulkUpdate(indexName, vulnerabilities)
	require.NoError(t, err)

	err = client.BulkUpdate(indexName, []byte(bulkRequest))
	require.NoError(t, err)

	// wait for data to be present
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		searchResponse := searchAllVulnerabilities(c, client)
		assert.Equal(c, expectedDocumentCount, searchResponse.Hits.Total.Value)
	}, 10*time.Second, 500*time.Millisecond)
}

func searchAllVulnerabilities(c *assert.CollectT, client *Client) *SearchResponse[*Vulnerability] {
	responseBody, err := client.Search(indexName, []byte(``))
	require.NoError(c, err)

	searchResponse, err := UnmarshalSearchResponse[*Vulnerability](responseBody)
	require.NoError(c, err)
	return searchResponse
}
