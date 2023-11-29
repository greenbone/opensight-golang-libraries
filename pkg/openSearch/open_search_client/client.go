// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_client

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/opensearch-project/opensearch-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type IndexFunctionsInterface interface {
	CreateIndex(indexName string, body io.Reader) error
}

//go:generate mockery --name OpenSearchClientWithIndex --with-expecter=true
type OpenSearchClientWithIndex[T Identifiable] interface {
	// Save save
	Save(document T) error

	// SearchOne search one
	SearchOne(body Json) (*T, error)

	// Search search
	Search(body Json) (*SearchResponse[T], error)

	// UpdateById update by id
	UpdateById(id string, body map[string]any) error

	// AsyncDeleteByQuery async delete by query
	AsyncDeleteByQuery(body Json) error

	// DeleteByQuery delete by query
	DeleteByQuery(body Json) error

	// DeleteById delete by id
	DeleteById(id string) error

	// SearchString
	SearchString(body string) (*SearchResponse[T], error)

	// GetClient
	GetClient() *opensearch.Client

	// SaveVulnerabilitiesToIndex
	SaveToIndex(indexName string, documents []T) error
}

type openSearchClient[T Identifiable] struct {
	client    *opensearch.Client
	indexName string
}

func NewOpenSearchClient[T Identifiable](initClient *opensearch.Client, defaultIndexName string) *openSearchClient[T] {
	return &openSearchClient[T]{
		client:    initClient,
		indexName: defaultIndexName,
	}
}

func (o *openSearchClient[T]) GetClient() *opensearch.Client {
	return o.client
}

func (o *openSearchClient[T]) UpdateById(id string, updateBody map[string]any) error {
	jsonString, err := jsoniter.Marshal(&updateBody)
	if err != nil {
		return errors.Wrapf(err, "openSearch client can't marshal body with ID %q to JSON for update", id)
	}

	updateRequestString := `{ "doc": ` + string(jsonString) + `}`

	updateResponse, err := o.client.Update(
		o.indexName, id, strings.NewReader(updateRequestString),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(updateResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(updateResponse.StatusCode, resultString, o.indexName)
}

func (o *openSearchClient[T]) Save(document T) error {
	documentJson, err := jsoniter.Marshal(document)
	if err != nil {
		return errors.WithStack(err)
	}

	insertResponse, err := o.client.Index(
		o.indexName,
		strings.NewReader(string(documentJson)),
		o.client.Index.WithRefresh("true"),
		o.client.Index.WithDocumentID(document.GetId()),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(insertResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	err = GetResponseError(insertResponse.StatusCode, resultString, o.indexName)
	if err != nil {
		return errors.WithStack(err)
	}

	response := &CreatedResponse{}
	err = jsoniter.Unmarshal(resultString, response)
	if err != nil {
		return errors.WithStack(err)
	}

	document.SetId(response.Id)

	return nil
}

func (o *openSearchClient[T]) SearchOne(body Json) (*T, error) {
	json, err := body.ToJson()
	if err != nil {
		return nil, err
	}

	return o.searchOne(json)
}

func (o *openSearchClient[T]) searchOne(jsonString string) (*T, error) {
	elements, err := o.search(jsonString)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(elements.Hits.SearchHits) == 0 {
		return nil, NewOpenSearchResourceNotFoundWithStack(
			fmt.Sprintf("Document not found '%s'", jsonString))
	}

	return &elements.Hits.SearchHits[0].Content, nil
}

func (o *openSearchClient[T]) SearchString(body string) (*SearchResponse[T], error) {
	return o.search(body)
}

func (o *openSearchClient[T]) Search(body Json) (*SearchResponse[T], error) {
	json, err := body.ToJson()
	if err != nil {
		return nil, err
	}
	return o.search(json)
}

func (o *openSearchClient[T]) search(body string) (*SearchResponse[T], error) {
	log.Trace().Msgf("search query - json:'%s'", body)

	searchResponse, err := o.client.Search(
		o.client.Search.WithIndex(o.indexName),
		o.client.Search.WithBody(strings.NewReader(body)),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resultString, err := io.ReadAll(searchResponse.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	log.Trace().Msgf("search response - statusCode:'%d' json:'%s'", searchResponse.StatusCode, resultString)

	err = GetResponseError(searchResponse.StatusCode, resultString, o.indexName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var results SearchResponse[T]

	err = jsoniter.Unmarshal(resultString, &results)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, searchHit := range results.Hits.SearchHits {
		searchHit.Content.SetId(searchHit.Id)
	}

	return &results, nil
}

func (o *openSearchClient[T]) SaveAll(documents []T) error {
	if len(documents) == 0 {
		return nil
	}

	var body strings.Builder
	body.Reset()

	for _, document := range documents {
		if document.GetId() != "" {
			body.WriteString(fmt.Sprintf(`{"index": { "_index" : "%s", "_id": "%s"}}`,
				o.indexName, document.GetId()) + "\n")
		} else {
			body.WriteString(fmt.Sprintf(`{"index": { "_index" : "%s"}}`,
				o.indexName) + "\n")
		}

		documentJson, err := jsoniter.Marshal(document)
		if err != nil {
			return errors.WithStack(err)
		}
		body.WriteString(string(documentJson) + "\n")
	}

	insertResponse, err := o.client.Bulk(
		strings.NewReader(body.String()),
		o.client.Bulk.WithIndex(o.indexName),
		o.client.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(insertResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(insertResponse.StatusCode, resultString, o.indexName)
}

func (o *openSearchClient[T]) AsyncDeleteByQuery(body Json) error {
	return deleteByQuery(body, true, o)
}

func (o *openSearchClient[T]) DeleteByQuery(body Json) error {
	return deleteByQuery(body, false, o)
}

// deleteByQuery deletes documents by a query
func deleteByQuery[T Identifiable](body Json, isAsync bool, o *openSearchClient[T]) error {
	json, err := body.ToJson()
	if err != nil {
		return err
	}

	deleteResponse, err := o.client.DeleteByQuery(
		[]string{o.indexName},
		strings.NewReader(json),
		o.client.DeleteByQuery.WithWaitForCompletion(!isAsync),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(deleteResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(deleteResponse.StatusCode, resultString, o.indexName)
}

// DeleteById deletes a document by id
func (o *openSearchClient[T]) DeleteById(id string) error {
	deleteResponse, err := o.client.Delete(
		o.indexName,
		id,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(deleteResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(deleteResponse.StatusCode, resultString, o.indexName)
}

func (o *openSearchClient[T]) SaveToIndex(indexName string, documents []T) error {
	if len(documents) == 0 {
		return nil
	}

	var body strings.Builder
	body.Reset()

	for _, document := range documents {
		body.WriteString(fmt.Sprintf(`{"index": { "_index" : "%s"}}`,
			indexName) + "\n")
		documentJson, err := jsoniter.Marshal(document)
		if err != nil {
			return errors.WithStack(err)
		}
		body.WriteString(string(documentJson) + "\n")
	}

	insertResponse, err := o.client.Bulk(
		strings.NewReader(body.String()),
		o.GetClient().Bulk.WithIndex(indexName),
		o.GetClient().Bulk.WithRefresh("true"),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	resultString, err := io.ReadAll(insertResponse.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetResponseError(insertResponse.StatusCode, resultString, indexName)
}

// GetResponseError gets the error from the response
func GetResponseError(statusCode int, responseString []byte, indexName string) error {
	if statusCode >= 200 && statusCode < 300 {
		errorResponse := &BulkResponse{}
		err := jsoniter.Unmarshal(responseString, errorResponse)
		if err != nil {
			return errors.WithStack(err)
		}

		if errorResponse.HasError {
			return errors.Errorf("request error %v", errorResponse)
		}

		return nil
	}

	if statusCode == http.StatusBadRequest {
		openSearchErrorResponse := &OpenSearchErrorResponse{}
		err := jsoniter.Unmarshal(responseString, openSearchErrorResponse)
		if err != nil {
			return errors.WithStack(err)
		}

		if openSearchErrorResponse.Error.Type == "resource_already_exists_exception" {
			return NewOpenSearchResourceAlreadyExistsWithStack(
				fmt.Sprintf("Resource '%s' already exists", indexName))
		} else {
			return NewOpenSearchErrorWithStack(openSearchErrorResponse.Error.Reason)
		}
	} else {
		return NewOpenSearchErrorWithStack(string(responseString))
	}
}
