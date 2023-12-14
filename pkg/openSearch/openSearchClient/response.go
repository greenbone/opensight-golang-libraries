// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type SearchResponseHit[T any] struct {
	Id      string `json:"_id"`
	Type    string `json:"_type"`
	Content T      `json:"_source"`
}

type SearchResponseHits[T any] struct {
	Total      SearchResponseHitsTotal
	SearchHits []SearchResponseHit[T] `json:"hits"`
}

type KeepJsonAsString []byte

func (k *KeepJsonAsString) UnmarshalJSON(data []byte) error {
	*k = data

	return nil
}

type DynamicAggregationHits struct {
	Total      SearchResponseHitsTotal `json:"total"`
	SearchHits KeepJsonAsString        `json:"hits"`
}

func UnmarshalSearchResponse[T any](data []byte) (*SearchResponse[T], error) {
	var results SearchResponse[T]

	if err := jsoniter.Unmarshal(data, &results); err != nil {
		return nil, errors.WithStack(err)
	}
	return &results, nil
}

type SearchResponseHitsTotal struct {
	Value    uint
	Relation string
}

type Bucket struct {
	Key         any    `json:"key"`
	KeyAsString string `json:"key_as_string"`
	DocCount    uint   `json:"doc_count"`
	Aggs        map[string]DynamicAggregation
}

func addToMapAsDynamicAggregation(
	iterator *jsoniter.Iterator, fieldName string, callbackExtra any,
) {
	aggMap := callbackExtra.(map[string]DynamicAggregation)

	if iterator.WhatIsNext() == jsoniter.ObjectValue {
		newAggregation := DynamicAggregation{}
		iterator.ReadVal(&newAggregation)
		aggMap[fieldName] = newAggregation
	} else {
		iterator.Skip()
	}
}

func (bucket *Bucket) UnmarshalJSON(bytes []byte) error {
	type _Bucket Bucket
	unmarshalBucket := _Bucket{}

	if err := Unmarshal(bytes, &unmarshalBucket); err != nil {
		return err
	}
	*bucket = Bucket(unmarshalBucket)
	bucket.Aggs = make(map[string]DynamicAggregation)
	parseUnknownFields(bytes, bucket, addToMapAsDynamicAggregation, bucket.Aggs)

	return nil
}

type DynamicAggregation struct {
	DocCountErrorUpperBound int                    `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        uint                   `json:"sum_other_doc_count"`
	Buckets                 []Bucket               `json:"buckets"`
	Value                   any                    `json:"value"`
	ValueAsString           any                    `json:"value_as_string"`
	Hits                    DynamicAggregationHits `json:"hits"`
}

type SearchResponseAggregation struct {
	DocCountErrorUpperBound int      `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        uint     `json:"sum_other_doc_count"`
	Buckets                 []Bucket `json:"buckets"`
	Value                   uint64   `json:"value"`
}

type SearchResponseAggregations map[string]SearchResponseAggregation

type SearchResponse[T any] struct {
	Took         uint                       `json:"took"`
	TimedOut     bool                       `json:"timed_out"`
	Hits         SearchResponseHits[T]      `json:"hits"`
	Aggregations SearchResponseAggregations `json:"aggregations"`
}

func (s SearchResponse[T]) GetSearchHits() []SearchResponseHit[T] {
	return s.Hits.SearchHits
}

// GetResults returns list of documents
func (s SearchResponse[T]) GetResults() []T {
	var results []T
	for _, hit := range s.Hits.SearchHits {
		results = append(results, hit.Content)
	}
	return results
}

type CreatedResponse struct {
	Id     string `json:"_id"`
	Result string `json:"result"`
}

// BulkResponse bulk response
type BulkResponse struct {
	Took     uint         `json:"took"`
	HasError bool         `json:"errors"`
	Errors   []IndexError `json:"items"`
}
