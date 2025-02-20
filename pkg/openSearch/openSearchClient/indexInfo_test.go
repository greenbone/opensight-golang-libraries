package openSearchClient

import (
	"testing"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/assert"
)

func TestConvertToIndexInfo(t *testing.T) {
	cases := []struct {
		name     string
		input    []opensearchapi.CatIndexResp
		expected []IndexInfo
	}{
		{
			name: "converts single index",
			input: []opensearchapi.CatIndexResp{
				{Index: "index1", CreationDate: 1625097600},
			},
			expected: []IndexInfo{
				{Name: "index1", CreationDate: 1625097600},
			},
		},
		{
			name: "converts multiple indices",
			input: []opensearchapi.CatIndexResp{
				{Index: "index1", CreationDate: 1625097600},
				{Index: "index2", CreationDate: 1625184000},
			},
			expected: []IndexInfo{
				{Name: "index1", CreationDate: 1625097600},
				{Name: "index2", CreationDate: 1625184000},
			},
		},
		{
			name:     "converts empty slice",
			input:    []opensearchapi.CatIndexResp{},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertToIndexInfo(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSortIndexInfoByCreationDate(t *testing.T) {
	cases := []struct {
		name     string
		input    []IndexInfo
		expected []IndexInfo
	}{
		{
			name: "sorts indices by creation date",
			input: []IndexInfo{
				{Name: "index2", CreationDate: 1625184000},
				{Name: "index1", CreationDate: 1625097600},
			},
			expected: []IndexInfo{
				{Name: "index1", CreationDate: 1625097600},
				{Name: "index2", CreationDate: 1625184000},
			},
		},
		{
			name: "handles already sorted indices",
			input: []IndexInfo{
				{Name: "index1", CreationDate: 1625097600},
				{Name: "index2", CreationDate: 1625184000},
			},
			expected: []IndexInfo{
				{Name: "index1", CreationDate: 1625097600},
				{Name: "index2", CreationDate: 1625184000},
			},
		},
		{
			name:     "handles empty slice",
			input:    []IndexInfo{},
			expected: []IndexInfo{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := SortIndexInfoByCreationDate(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
