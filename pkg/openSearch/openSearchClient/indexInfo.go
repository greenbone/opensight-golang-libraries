// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"sort"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type IndexInfo struct {
	Name         string
	CreationDate int // Store Unix timestamp
}

func ConvertToIndexInfo(indices []opensearchapi.CatIndexResp) []IndexInfo {
	var indexes []IndexInfo
	for _, indexInfo := range indices {
		indexes = append(indexes, IndexInfo{
			Name:         indexInfo.Index,
			CreationDate: indexInfo.CreationDate,
		})
	}
	return indexes
}

type ByCreationDate []IndexInfo

func (a ByCreationDate) Len() int           { return len(a) }
func (a ByCreationDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreationDate) Less(i, j int) bool { return a[i].CreationDate < a[j].CreationDate }

func SortIndexInfoByCreationDate(indexes []IndexInfo) []IndexInfo {
	sort.Sort(ByCreationDate(indexes))
	return indexes
}
