// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

import (
	"strings"

	"github.com/aquasecurity/esquery"
)

// NestedQuery represents an OpenSearch nested query.
type NestedQuery struct {
	Path  string            `json:"path"`
	Query esquery.BoolQuery `json:"query"`
}

// Nested creates a new NestedQuery.
func Nested(field string, q esquery.BoolQuery) *NestedQuery {
	return &NestedQuery{
		Path:  calculatePath(field),
		Query: q,
	}
}

func calculatePath(field string) string {
	parts := strings.Split(field, ".")
	if len(parts) < 2 {
		return "" // return empty or handle error accordingly
	}
	return parts[0] + "." + parts[1]
}

// Map returns a map representation of the NestedQuery, thus implementing the esquery.Mappable interface.
// Used for serialization to JSON.
func (nq *NestedQuery) Map() map[string]interface{} {
	return map[string]interface{}{
		"nested": map[string]interface{}{
			"path":  nq.Path,
			"query": nq.Query.Map(), // since BoolQuery implements Mappable
		},
	}
}
