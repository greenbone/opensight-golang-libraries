// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

// MatchQuery represents an OpenSearch match part in an OpenSearch query as described in
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-filter-context.html#query-filter-context-ex
type MatchQuery struct {
	Field string
	Value interface{}
}

// Map returns a map representation of the MatchQuery, thus implementing the esquery.Mappable interface.
// Used for serialization to JSON.
func (mq *MatchQuery) Map() map[string]interface{} {
	return map[string]interface{}{
		"match": map[string]interface{}{
			mq.Field: mq.Value,
		},
	}
}

// Match creates a new MatchQuery.
func Match(field string, value interface{}) *MatchQuery {
	return &MatchQuery{
		Field: field,
		Value: value,
	}
}
