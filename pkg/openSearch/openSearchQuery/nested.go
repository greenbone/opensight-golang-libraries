package openSearchQuery

import (
	"strings"

	"github.com/aquasecurity/esquery"
)

type NestedQuery struct {
	Path  string            `json:"path"`
	Query esquery.BoolQuery `json:"query"`
}

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

func (nq *NestedQuery) Map() map[string]interface{} {
	return map[string]interface{}{
		"nested": map[string]interface{}{
			"path":  nq.Path,
			"query": nq.Query.Map(), // since BoolQuery implements Mappable
		},
	}
}
