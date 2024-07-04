// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"fmt"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"

	"github.com/aquasecurity/esquery"
)

type effectiveSortField struct {
	plainField       *string
	aggregationName  *string
	aggregationValue string
}

var sortFieldMapping = map[string]effectiveSortField{
	"severity": {
		plainField:       strPtr("vulnerabilityTest.severityCvss.override"),
		aggregationName:  strPtr("maxSeverity"),
		aggregationValue: "maxSeverity.value",
	},
	"qod": {
		plainField:       strPtr("qod"),
		aggregationName:  strPtr("maxQod"),
		aggregationValue: "maxQod.value",
	},
	"assetCount": {
		plainField:       nil,
		aggregationName:  nil,
		aggregationValue: "uniqueAssetCount.value",
	},
}

func strPtr(s string) *string {
	return &s
}

func AddOrder(aggregation *esquery.TermsAggregation, sortingRequest *sorting.Request) (*esquery.TermsAggregation, error) {
	if sortingRequest != nil {
		field, err := effectiveSortFieldOf(*sortingRequest)
		if err != nil {
			return nil, err
		}
		return aggregation.Order(map[string]string{field.aggregationValue: sortingRequest.SortDirection.String()}), nil
	}
	return aggregation, nil
}

func AddMaxAggForSorting(aggs []esquery.Aggregation, sortingRequest *sorting.Request) ([]esquery.Aggregation, error) {
	if sortingRequest == nil {
		return aggs, nil
	}

	field, err := effectiveSortFieldOf(*sortingRequest)
	if err != nil {
		return nil, err
	}

	// refers to an existing aggregation that does not need to be created
	if field.plainField == nil || field.aggregationName == nil {
		return aggs, nil
	}

	agg := esquery.Max(*field.aggregationName, *field.plainField)
	return append(aggs, agg), nil
}

// BucketSortAgg is capable to sort all existing buckets, but is currently only used for paging
func BucketSortAgg(sortingRequest *sorting.Request, pagingRequest *paging.Request) (*esquery.CustomAggMap, error) {
	if sortingRequest == nil && pagingRequest == nil {
		return nil, nil
	}

	sorting := map[string]interface{}{}

	if sortingRequest != nil {
		field, err := effectiveSortFieldOf(*sortingRequest)
		if err != nil {
			return nil, err
		}

		order, err := getOrder(sortingRequest)
		if err != nil {
			return nil, err
		}

		sorting = map[string]interface{}{
			"sort": []map[string]interface{}{
				{field.aggregationValue: map[string]interface{}{
					"order": order,
				}},
			},
		}
	}

	if pagingRequest != nil {
		sorting["from"] = pagingRequest.PageIndex * pagingRequest.PageSize
		sorting["size"] = pagingRequest.PageSize
	}

	agg := esquery.CustomAgg("sorting", map[string]interface{}{
		"bucket_sort": sorting,
	})
	return agg, nil
}

func AddBucketSortAgg(aggs []esquery.Aggregation, sortingRequest *sorting.Request,
	pagingRequest *paging.Request,
) ([]esquery.Aggregation, error) {
	agg, err := BucketSortAgg(sortingRequest, pagingRequest)
	if err != nil {
		return nil, err
	}

	if agg == nil {
		return aggs, nil
	}

	return append(aggs, agg), nil
}

func getOrder(sortingRequest *sorting.Request) (esquery.Order, error) {
	switch strings.ToLower(sortingRequest.SortDirection.String()) {
	case string(esquery.OrderAsc):
		return esquery.OrderAsc, nil
	case string(esquery.OrderDesc):
		return esquery.OrderDesc, nil
	default:
		return "", fmt.Errorf("%s is no valid sort direction", sortingRequest.SortDirection.String())
	}
}

func effectiveSortFieldOf(sortingRequest sorting.Request) (effectiveSortField, error) {
	field, ok := sortFieldMapping[sortingRequest.SortColumn]

	if !ok {
		return effectiveSortField{}, fmt.Errorf("%s is no valid sort column, possible values: %s",
			sortingRequest.SortColumn, strings.Join(validSortColumns(), ","))
	}
	return field, nil
}

// TODO replace with generic function
func validSortColumns() []string {
	validSortColumns := make([]string, 0, len(sortFieldMapping))
	for k := range sortFieldMapping {
		validSortColumns = append(validSortColumns, k)
	}
	return validSortColumns
}
