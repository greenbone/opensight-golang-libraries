// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

// TermsSource represents a terms value source to composite aggregations, as described in
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/search-aggregations-bucket-composite-aggregation.html#_terms
// see also CompositeAgg
type TermsSource struct {
	name  string
	field string
}

func Terms(name string, field string) *TermsSource {
	return &TermsSource{
		name:  name,
		field: field,
	}
}

func (t *TermsSource) Map() map[string]interface{} {
	return map[string]interface{}{
		t.name: map[string]interface{}{
			"terms": map[string]interface{}{
				"field": t.field,
			},
		},
	}
}
