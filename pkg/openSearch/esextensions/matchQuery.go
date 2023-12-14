package esextensions

// MatchQuery represents an OpenSearch match query.
type MatchQuery struct {
	Field string
	Value interface{}
}

const useSimpleMap = true

// Map returns a map representation of the MatchQuery, thus implementing the esquery.Mappable interface.
// Used for serialization to JSON.
func (mq *MatchQuery) Map() map[string]interface{} {
	if useSimpleMap {
		return map[string]interface{}{
			"match": map[string]interface{}{
				mq.Field: mq.Value,
			},
		}
	} else {
		return map[string]interface{}{
			"match": map[string]interface{}{
				mq.Field: map[string]interface{}{
					"query": mq.Value,
				},
			},
		}
	}
}

// Match creates a new MatchQuery.
func Match(field string, value interface{}) *MatchQuery {
	return &MatchQuery{
		Field: field,
		Value: value,
	}
}
