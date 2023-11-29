// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_client

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchResponseUnmarshal(t *testing.T) {
	json := `{
	  "aggregations": {
		"hostnames": {
		  "doc_count_error_upper_bound": -1,
		  "sum_other_doc_count": 2,
		  "buckets": [
			{
			  "key": "testKey",
			  "doc_count": 1280,
			  "agg": {
				"doc_count_error_upper_bound": 3,
				"sum_other_doc_count": 4,
				"buckets": [
				  {
					"key": "5",
					"doc_count": 6,
					"inner_agg": {
                      "value": 7.0
					}
				  }
				]
			  }
			}
		  ]
		}
	  }
	}`

	var results SearchResponse[Identifiable]
	err := jsoniter.Unmarshal([]byte(json), &results)
	require.NoError(t, err)

	hostnames := results.Aggregations["hostnames"]

	assert.Equal(t, int(-1), hostnames.DocCountErrorUpperBound)
	assert.Equal(t, uint(2), hostnames.SumOtherDocCount)
	assert.Equal(t, "testKey", hostnames.Buckets[0].Key)
	assert.Equal(t, uint(1280), hostnames.Buckets[0].DocCount)
	assert.NotNil(t, hostnames.Buckets[0].Aggs["agg"])
	assert.Equal(t, int(3), hostnames.Buckets[0].Aggs["agg"].DocCountErrorUpperBound)
	assert.Equal(t, uint(4), hostnames.Buckets[0].Aggs["agg"].SumOtherDocCount)
	assert.NotNil(t, hostnames.Buckets[0].Aggs["agg"].Buckets[0])
	assert.Equal(t, "5", hostnames.Buckets[0].Aggs["agg"].Buckets[0].Key)
	assert.Equal(t, uint(6), hostnames.Buckets[0].Aggs["agg"].Buckets[0].DocCount)
	assert.Equal(t, 7.0, hostnames.Buckets[0].Aggs["agg"].Buckets[0].Aggs["inner_agg"].Value)
}
