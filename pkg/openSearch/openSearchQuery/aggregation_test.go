// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregation(t *testing.T) {
	t.Run("shouldCreateJsonForNestedAggregations", func(t *testing.T) {
		query := NewBoolQueryBuilder(&QuerySettings{})
		query.AddAggregation(
			TermsAgg("aggNameOne", "hostname").
				Aggs(
					TermsAgg("aggNameTwo", "host").Aggs(
						ValueCountAgg("aggNameThree", "id"),
					),
				),
		)

		json, err := query.ToJson()
		require.NoError(t, err)

		assert.JSONEq(t, `{
		  "aggs": {
			"aggNameOne": {
			  "aggs": {
				"aggNameTwo": {
				  "terms": {
					"field": "host"
				  },
                  "aggs": {
					"aggNameThree": {
						"value_count": {
							"field": "id"
						}
					}
				  }
				}
			  },
			  "terms": {
				"field": "hostname"
			  }
			}
		  },
		  "query": {
			"bool": {}
		  },
		  "size": 100
		}`, json)
	})
}
