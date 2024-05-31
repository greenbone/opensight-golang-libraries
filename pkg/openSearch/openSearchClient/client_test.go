// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"context"
	config2 "github.com/greenbone/opensight-golang-libraries/pkg/openSearch/openSearchClient/config"
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

type MockTokenReceiver struct {
}

func (m *MockTokenReceiver) GetClientAccessToken(clientName, clientSecret string) (string, error) {
	return "test-token", nil
}

func TestClient(t *testing.T) {
	type testCase struct {
		testFunc func(t *testing.T, client *Client)
	}
	tcs := map[string]testCase{
		"TestBulkUpdate": {func(t *testing.T, client *Client) {
			// given
			bulkRequest, err := SerializeDocumentsForBulkUpdate(indexName, []*Vulnerability{&aVulnerability})
			require.NoError(t, err)

			// when
			err = client.BulkUpdate(indexName, bulkRequest)

			// then
			require.NoError(t, err)
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				searchResponse := searchAllVulnerabilities(c, client)
				assert.Equal(c, uint(1), searchResponse.Hits.Total.Value)
				assert.Equal(c, 1, len(searchResponse.GetResults()))
				assert.Equal(c, aVulnerability, *searchResponse.GetResults()[0])
			}, 10*time.Second, 500*time.Millisecond)
		}},
		"TestUpdate": {func(t *testing.T, client *Client) {
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
		"TestAsyncDeleteByQuery": {func(t *testing.T, client *Client) {
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
		"TestDeleteByQuery": {func(t *testing.T, client *Client) {
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
		"TestSearch": {func(t *testing.T, client *Client) {
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
	}

	ctx := context.Background()
	opensearchContainer, conf, err := StartOpensearchTestContainer(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, opensearchContainer)
	defer func() {
		opensearchContainer.Terminate(ctx)
	}()

	opensearchProjectClient, err := NewOpenSearchProjectClient(context.Background(), conf)
	require.NoError(t, err)
	require.NotNil(t, opensearchProjectClient)

	iFunc := NewIndexFunction(opensearchProjectClient)
	config := config2.OpensearchClientConfig{
		UpdateMaxRetries: 1,
		UpdateRetrySleep: 1,
	}
	mockTokenReceiver := MockTokenReceiver{}
	client := NewClient(opensearchProjectClient, config, &mockTokenReceiver)

	for testName, testCase := range tcs {
		t.Run(testName, func(t *testing.T) {
			err = iFunc.DeleteIndex(indexName)
			if err != nil {
				log.Trace().Msgf("Error while deleting index: %s", err)
			}
			schema := folder.GetContent(t, "testdata/testSchema.json")
			err = iFunc.CreateIndex(indexName, []byte(schema))
			require.NoError(t, err)
			testCase.testFunc(t, client)
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

func TestGetResponseError(t *testing.T) {
	//given
	testCases := map[string]struct {
		givenStatusCode      int
		givenResponse        string
		expectedErrorMessage *string
	}{
		"no error": {
			givenStatusCode:      200,
			givenResponse:        `{"took": 1, "errors": false, "items": []}`,
			expectedErrorMessage: nil,
		},
		"broken json": {
			givenStatusCode:      200,
			givenResponse:        `{"error": {`,
			expectedErrorMessage: strPtr(`expect " after {`),
		},
		"error in body": {
			givenStatusCode:      200,
			givenResponse:        `{"took": 1, "errors": true, "items": [{"index": {}}]}`,
			expectedErrorMessage: strPtr(`{"took": 1, "errors": true, "items": [{"index": {}}]}`),
		},
		"index exists": {
			givenStatusCode: 400,
			givenResponse: `{
  "error": {
    "root_cause": [
      {
        "type": "resource_already_exists_exception",
        "reason": "index [test/CafPswH_Q5mGXcKgBN3TNg] already exists",
        "index": "test",
        "index_uuid": "CafPswH_Q5mGXcKgBN3TNg"
      }
    ],
    "type": "resource_already_exists_exception",
    "reason": "index [test/CafPswH_Q5mGXcKgBN3TNg] already exists",
    "index": "test",
    "index_uuid": "CafPswH_Q5mGXcKgBN3TNg"
  },
  "status": 400
}`,
			expectedErrorMessage: strPtr(`Resource 'test' already exists`),
		},
		"bad request with broken json": {
			givenStatusCode:      400,
			givenResponse:        `{"error": {`,
			expectedErrorMessage: strPtr(`openSearchClient.OpenSearchErrors.ReadString`),
		},
		"some other bad request": {
			givenStatusCode: 400,
			givenResponse: `{
  "error": {
    "type": "some_bad_request",
    "reason": "some reason",
    "index": "test",
    "index_uuid": "CafPswH_Q5mGXcKgBN3TNg"
  },
  "status": 400
}`,
			expectedErrorMessage: strPtr(`some reason`),
		},
		"internal server error": {
			givenStatusCode:      500,
			givenResponse:        `some error message`,
			expectedErrorMessage: strPtr(`some error message`),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// when
			err := GetResponseError(tc.givenStatusCode, []byte(tc.givenResponse), indexName)

			// then
			if tc.expectedErrorMessage != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), *tc.expectedErrorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
