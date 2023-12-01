package query

import (
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
)

// ResponseListWithMetadata represents a list of responses including the filter and paging metadata.
type ResponseListWithMetadata[T any] struct {
	Metadata Metadata `json:"metadata" binding:"required"`
	Data     []T      `json:"data" binding:"required"`
}

type ResponseWithMetadata[T any] struct {
	Metadata Metadata `json:"metadata" binding:"required"`
	Data     T        `json:"data" binding:"required"`
}

type Metadata struct {
	Filter  *filter.Request  `json:"filter" binding:"required"`
	Paging  *paging.Response `json:"paging,omitempty"`
	Sorting *sorting.Request `json:"sorting,omitempty"`
}

func NewMetadata(resultSelector ResultSelector, totalRowCount uint64) Metadata {
	for i, field := range resultSelector.Filter.Fields {
		if field.Keys == nil {
			resultSelector.Filter.Fields[i].Keys = []string{}
		}
	}
	return Metadata{
		Filter:  resultSelector.Filter,
		Paging:  paging.NewResponse(resultSelector.Paging, totalRowCount),
		Sorting: resultSelector.Sorting,
	}
}
