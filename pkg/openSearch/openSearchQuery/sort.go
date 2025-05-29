// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
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

type EffectiveSortField struct {
	PlainField       *string
	AggregationName  *string
	AggregationValue string
}

func AddOrder(aggregation *esquery.TermsAggregation, sortingRequest *sorting.Request,
	sortFieldMapping map[string]EffectiveSortField,
) (*esquery.TermsAggregation, error) {
	if sortingRequest != nil {
		field, err := effectiveSortFieldOf(*sortingRequest, sortFieldMapping)
		if err != nil {
			return nil, err
		}
		return aggregation.Order(map[string]string{field.AggregationValue: sortingRequest.SortDirection.String()}), nil
	}
	return aggregation, nil
}

func AddMaxAggForSorting(aggs []esquery.Aggregation, sortingRequest *sorting.Request,
	sortFieldMapping map[string]EffectiveSortField,
) ([]esquery.Aggregation, error) {
	if sortingRequest == nil {
		return aggs, nil
	}

	field, err := effectiveSortFieldOf(*sortingRequest, sortFieldMapping)
	if err != nil {
		return nil, err
	}

	// refers to an existing aggregation that does not need to be created
	if field.PlainField == nil || field.AggregationName == nil {
		return aggs, nil
	}

	agg := esquery.Max(*field.AggregationName, *field.PlainField)
	return append(aggs, agg), nil
}

// BucketSortAgg is capable to sort all existing buckets, but is currently only used for paging
func BucketSortAgg(sortingRequest *sorting.Request, sortFieldMapping map[string]EffectiveSortField,
	pagingRequest *paging.Request,
) (*esquery.CustomAggMap, error) {
	if sortingRequest == nil && pagingRequest == nil {
		return nil, nil
	}

	sorting := map[string]interface{}{}

	if sortingRequest != nil {
		field, err := effectiveSortFieldOf(*sortingRequest, sortFieldMapping)
		if err != nil {
			return nil, err
		}

		order, err := getOrder(sortingRequest)
		if err != nil {
			return nil, err
		}

		sorting = map[string]interface{}{
			"sort": []map[string]interface{}{
				{field.AggregationValue: map[string]interface{}{
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
	sortFieldMapping map[string]EffectiveSortField,
	pagingRequest *paging.Request,
) ([]esquery.Aggregation, error) {
	agg, err := BucketSortAgg(sortingRequest, sortFieldMapping, pagingRequest)
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

func effectiveSortFieldOf(sortingRequest sorting.Request, sortFieldMapping map[string]EffectiveSortField) (EffectiveSortField, error) {
	field, ok := sortFieldMapping[sortingRequest.SortColumn]

	if !ok {
		return EffectiveSortField{}, fmt.Errorf("%s is no valid sort column, possible values: %s",
			sortingRequest.SortColumn, strings.Join(validSortColumns(sortFieldMapping), ","))
	}
	return field, nil
}

// TODO replace with generic function
func validSortColumns(sortFieldMapping map[string]EffectiveSortField) []string {
	validSortColumns := make([]string, 0, len(sortFieldMapping))
	for k := range sortFieldMapping {
		validSortColumns = append(validSortColumns, k)
	}
	return validSortColumns
}
