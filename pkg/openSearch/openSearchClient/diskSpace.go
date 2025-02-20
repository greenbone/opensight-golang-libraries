package openSearchClient

import (
	"context"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type OpenSearchHealth struct {
	openSearchProjectClient *opensearchapi.Client
	context                 context.Context
}

func NewOpenSearchHealth(openSearchProjectClient *opensearchapi.Client) *OpenSearchHealth {
	return &OpenSearchHealth{openSearchProjectClient: openSearchProjectClient}
}

func (h *OpenSearchHealth) GetDiskAllocationPercentage() (int, error) {
	request := opensearchapi.CatAllocationReq{
		Params: opensearchapi.CatAllocationParams{
			Bytes:  "gb",
			Pretty: true,
		},
	}
	allocation, err := h.openSearchProjectClient.Cat.Allocation(h.context, &request)
	if err != nil {
		return 0, err
	}
	diskPercent := allocation.Allocations[0].DiskPercent
	return lo.FromPtr(diskPercent), nil
}

func (h *OpenSearchHealth) GetIndexesWithCreationDate(pattern string) ([]IndexInfo, error) {

	request := opensearchapi.CatIndicesReq{
		Indices: []string{pattern},
		Params: opensearchapi.CatIndicesParams{
			H: []string{"index,creation.date"},
		},
	}
	response, err := h.openSearchProjectClient.Cat.Indices(h.context, &request)
	if err != nil {
		log.Debug().Err(err).Msg("error while checking if index exists")
		return nil, err
	}
	indices := response.Indices
	indexInfos := ConvertToIndexInfo(indices)
	return indexInfos, nil
}
