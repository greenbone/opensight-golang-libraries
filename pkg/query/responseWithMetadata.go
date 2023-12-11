package query

import (
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
)

// ResponseListWithMetadata represents a response containing a list of data with associated metadata.
// The 'Metadata' field is of type 'Metadata' includes filter, paging, and sorting information used in the query.
// The 'Data' field is a slice of type 'T' and represents the data retrieved.
type ResponseListWithMetadata[T any] struct {
	Metadata Metadata `json:"metadata" binding:"required"`
	Data     []T      `json:"data" binding:"required"`
}

// ResponseWithMetadata represents a response with associated metadata.
// The metadata includes filter, paging, and sorting information.
// The 'Metadata' field is of type 'Metadata' includes filter, paging, and sorting information used in the query.
// The 'Data' field is of any type and represents the data retrieved by the query.
type ResponseWithMetadata[T any] struct {
	Metadata Metadata `json:"metadata" binding:"required"`
	Data     T        `json:"data" binding:"required"`
}

// Metadata represents the metadata used in a query.
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
