// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package ostesting

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

const (
	// defaultTestPassword used to connect to the client, same as hardcoded in ./tests/docker-compose.yml
	defaultTestPassword = "secureTestPassword444!"
	defaultTestAddress  = "https://localhost:9300"
	defaultTestUser     = "admin"
)

var (
	// defaultOSConf is the default configuration used for testing OpenSearch to connect
	// to the test OpenSearch instance running in docker
	defaultOSConf = config.OpenSearch{
		Address:             defaultTestAddress,
		SkipSSLVerification: true,
		User:                defaultTestUser,
		Password:            defaultTestPassword,
		AuthMethod:          config.BasicAuth,
	}
)

// Tester manages connecting with testing OpenSearch instance and implements
// helper methods
type Tester struct {
	t *testing.T

	osClient *opensearchapi.Client

	conf     config.OpenSearch
	parallel bool
}

type TesterOption func(tst *Tester)

// RunNotParallelOption signifies that tests associated with [Tester]
// should not be run in parallel (by default they are)
func RunNotParallelOption(tst *Tester) {
	tst.parallel = false
}

// WithAddress sets the custom address of testing opensearch
// instance to which the tester should point to
func WithAddress(address string) TesterOption {
	return func(tst *Tester) {
		tst.conf.Address = address
	}
}

// WithConfig is an option to use custom [config.OpenSearch] for tester.
func WithConfig(conf config.OpenSearch) TesterOption {
	return func(tst *Tester) {
		tst.conf = conf
	}
}

// NewTester initializes new Tester
// It runs tests associated with [t] as parallel by default, unless runNotParallel option
// is provided.
func NewTester(t *testing.T, opts ...TesterOption) *Tester {
	tst := &Tester{
		t:        t,
		parallel: true,
		conf:     defaultOSConf,
	}

	for _, opt := range opts {
		opt(tst)
	}

	osClient, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // not for production use
			},
			Addresses: []string{tst.conf.Address},
			Username:  tst.conf.User,
			Password:  tst.conf.Password,
		}})
	if err != nil {
		t.Fatalf("error while initializing opensearchapi.Client for testing: %v", err)
	}

	tst.osClient = osClient

	if tst.parallel {
		t.Parallel()
	}

	return tst
}

// OSClient returns [*opensearchapi.Client] associated with test OpenSearch instance
func (tst Tester) OSClient() *opensearchapi.Client {
	return tst.osClient
}

// T returns *testing.T associated with tester
func (tst Tester) T() *testing.T {
	return tst.t
}

// Config returns opensearch config that can be used to initialize
// client using test OpenSearch instance
func (tst Tester) Config() config.OpenSearch {
	return tst.conf
}

// NewIndex creates new index with unique name generated based on provided [prefix].
// It returns the generated index name.
// Note: in most cases the _IndexAlias functions should be used that create both index and add it
// to alias as this represents the typical usage of indices (accessing them through alias) in the code.
func (tst Tester) NewIndex(t *testing.T, prefix string, mapping *string) string {
	var body io.Reader
	if mapping != nil {
		body = strings.NewReader(*mapping)
	} else {
		body = strings.NewReader("")
	}

	// generate unique index name
	random, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}
	indexName := prefix + "_" + random.String()[0:8]

	createResponse, err := tst.osClient.Indices.Create(
		context.Background(),
		opensearchapi.IndicesCreateReq{
			Index: indexName,
			Body:  body,
		},
	)
	if err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	t.Logf("created index %s", indexName)

	tst.deleteIndexOnCleanup(t, indexName)

	var resp *opensearch.Response
	if createResponse != nil {
		resp = createResponse.Inspect().Response
	}
	defer func() { //close response body
		if resp != nil {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}
	}()

	return indexName
}

// NewTestTypeIndex creates new index appropriate to use with [testType] documents with given [prefix]
// of index name. Internally it calls [Tester.NewIndex].
func (tst Tester) NewTestTypeIndex(t *testing.T, prefix string) string {
	return tst.NewIndex(t, prefix, &testTypeMapping)
}

// NewIndexAlias creates new uniquely named index together with associated alias using mapping if not nil (otherwise with dynamic mapping).
// It also registers [t].Cleanup function that deletes the index after test.
//
// [prefix] is the prefix name of the index that would be used to generate new unique index and alias name.
// Generation of unique index name is there to allow tests running in parallel and not interfere with each other.
//
// Internally opensearchapi.Client is called directly rather than opensearch.Client, to avoid
// using tested object in testing setup and to not have to adjust opensearch.Client methods just for
// the sake of being used by Tester (as tester use-case of generating unique index/alias with
// custom mapping differs from production use-cases).
//
// Upon succesful creation the function returns index and associated alias names, in case of error [t] is used to mark test
// as failed.
// Note: usually the index should be accessed through alias name. In some cases the method needs
// to receive the concrete index instead.
func (tst Tester) NewIndexAlias(t *testing.T, prefix string, mapping *string) (index, alias string) {
	index = tst.NewIndex(t, prefix, mapping)
	alias = index + "_alias" // use generated unique [index] name in building the alias name
	tst.addIndexToAlias(t, index, alias)

	return index, alias
}

// NewTestTypeIndexAlias creates new index appropriate to use with [testType] documents
// with given index and alias names [prefix] and returns new index and alias names.
// Internally it calls [Tester.NewIndexAlias].
func (tst Tester) NewTestTypeIndexAlias(t *testing.T, prefix string) (string, string) {
	return tst.NewIndexAlias(t, prefix, &testTypeMapping)
}

// NewNamedIndexAlias works like [Tester.NewIndexAlias], except the name of the newly created index alias will be
// exactly as provided with [name] instead of being generated based on provided prefix.
// On successful creation it returns the concrete underlying index name
// It can be used in testing functionalities that rely on defined index alias name (eg. IndexVTS)
// and for which generated index alias name would not work.
// It goes with a drawback of having to make sure that other test run in parallel will not use
// the same index alias and interfere - while internal indices names are still unique, they would refer to
// the same alias.
func (tst Tester) NewNamedIndexAlias(t *testing.T, name string, mapping *string) string {
	index := tst.NewIndex(t, name, mapping)
	tst.addIndexToAlias(t, index, name)

	return index
}

// CreateDocuments creates documents [docs] on index [index] with IDs [ids]. If provided [ids] is nil
// the IDs for created documents will be generated.
func (tst Tester) CreateDocuments(t *testing.T, index string, docs []any, ids []string) {
	err := tst.CreateDocumentsReturningError(index, docs, ids)
	assert.NoError(t, err)
}

// CreateDocumentsReturningError creates documents [docs] on index [index] with IDs [ids]. If provided [ids] is nil
// the IDs for created documents will be generated.
// Instead of using [*testing.T] to fail on errors (like CreateDocuments) it returns
// error. Can be used when we expect error on creating documents in index and do not want
// to fail a test on it.
func (tst Tester) CreateDocumentsReturningError(index string, docs []any, ids []string) error {
	if ids != nil {
		if len(docs) != len(ids) {
			return fmt.Errorf("length of docs %v is not equal to length of ids %v", len(docs), len(ids))
		}
	} else {
		ids = make([]string, 0, len(docs))
		for i := 0; i < len(docs); i++ {
			ids = append(ids, uuid.NewString())
		}
	}

	for i, doc := range docs {
		b, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("could not marshal document: %w", err)
		}

		req := opensearchapi.DocumentCreateReq{
			Index:      index,
			Body:       bytes.NewReader(b),
			DocumentID: ids[i],
		}
		_, err = tst.osClient.Document.Create(
			context.Background(),
			req,
		)
		if err != nil {
			return fmt.Errorf("error while creating document on index %v: %w", index, err)
		}
	}

	if err := tst.refreshIndex(index); err != nil {
		return fmt.Errorf("error while flushing index %v: %w", index, err)
	}

	return nil
}

// RefreshIndex waits for index to refresh, so the updated documents can be obtained
func (tst Tester) RefreshIndex(t *testing.T, index string) {
	err := tst.refreshIndex(index)
	assert.NoError(t, err)
}

// GetTestTypeDocuments returns all documents of type [TestType] from [index]
func (tst Tester) GetTestTypeDocuments(t *testing.T, index string) []TestType {
	return GetDocuments[TestType](t, &tst, index)
}

// GetDocumentsReturningError returns all documents added to test index and returns
// error (instead of failing the tests like [GetDocuments] or [GetTestTypeDocuments])
func GetDocumentsReturningError[T any](tester *Tester, index string) ([]T, error) {
	reqBody := `{
    "query": {
        "match_all": {}
    }
}`
	searchResponse, err := tester.osClient.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{index},
			Body:    strings.NewReader(reqBody),
			Params: opensearchapi.SearchParams{
				TrackTotalHits: true,
			},
		},
	)
	var resp *opensearch.Response
	if searchResponse != nil {
		resp = searchResponse.Inspect().Response
	}

	defer func() { //close response body
		if resp != nil {
			if err := resp.Body.Close(); err != nil {
				log.Error().Msgf("failed to close response body: %v", err)
			}
		}
	}()

	if err != nil {
		return nil, fmt.Errorf("search call failed: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}

	docs, _, err := dtos.ParseOpenSearchListResponse[T](body)
	return docs, err
}

// GetDocuments returns all documents added to test index
func GetDocuments[T any](t *testing.T, tester *Tester, index string) []T {
	docs, err := GetDocumentsReturningError[T](tester, index)
	assert.NoError(t, err)
	return docs
}

// DeleteIndex deletes indices with given name (or pattern, eg. "index-name*")
func (tst Tester) DeleteIndex(t *testing.T, index string) {
	deleteResp, err := tst.osClient.Indices.Delete(context.Background(), opensearchapi.IndicesDeleteReq{
		Indices: []string{index},
	})
	var resp *opensearch.Response
	if deleteResp != nil {
		resp = deleteResp.Inspect().Response
	}
	if err != nil && resp != nil && resp.StatusCode == http.StatusNotFound {
		// do not throw error if index does not exist
		t.Logf("index %v supposed to be deleted on cleanup does not exist", index)
		return
	}
	assert.NoError(t, err)
	t.Logf("deleted test index %v", index)
}

func (tst Tester) addIndexToAlias(t *testing.T, index string, alias string) {
	reqBody := fmt.Sprintf(`{
      "actions": [
        {
          "add": {
            "index": "%[1]s",
            "alias": "%[2]s",
            "is_write_index": true
          }
        }
      ]
    }`, index, alias)
	req := opensearchapi.AliasesReq{
		Body: strings.NewReader(reqBody),
	}
	_, err := tst.osClient.Aliases(context.Background(), req)
	if err != nil {
		t.Fatalf("error while adding test index to alias: %v", err)
	}

	t.Logf("added index %s to alias %s", index, alias)
}

func (tst Tester) deleteIndexOnCleanup(t *testing.T, index string) {
	t.Cleanup(func() {
		if os.Getenv(testconfig.KeepFailedEnv) != "" {
			if t.Failed() {
				return
			}
		}
		tst.DeleteIndex(t, index)
	})
}

func (tst Tester) refreshIndex(index string) error {
	_, err := tst.osClient.Indices.Refresh(
		context.Background(),
		&opensearchapi.IndicesRefreshReq{
			Indices: []string{index},
		},
	)
	if err != nil {
		return fmt.Errorf("error while flushing index %v: %w", index, err)
	}

	return nil
}
