package esextensions

// MatchQuery represents an OpenSearch match query.
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
