package openSearchQuery

type MatchQuery struct {
	Field string
	Value interface{}
}

const useSimpleMap = true

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

func Match(field string, value interface{}) *MatchQuery {
	return &MatchQuery{
		Field: field,
		Value: value,
	}
}
