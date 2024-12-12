// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package esextensions

// TermsSource represents a terms value source in composite aggregations.
type TermsSource struct {
	name  string
	field string
	order string // Add order field for sorting
}

// Terms creates a new TermsSource.
//
// name: The name of the terms TermsSource.
// field: The name of the field referenced.
func Terms(name string, field string) *TermsSource {
	return &TermsSource{
		name:  name,
		field: field,
		order: "asc", // Default order is ascending
	}
}

// Order sets the sorting order for the TermsSource.
// Valid values: "asc", "desc".
func (t *TermsSource) Order(order string) *TermsSource {
	t.order = order
	return t
}

// Map returns a map representation of the TermsSource.
func (t *TermsSource) Map() map[string]interface{} {
	termsMap := map[string]interface{}{
		"field": t.field,
	}

	if t.order != "" {
		termsMap["order"] = t.order
	}

	return map[string]interface{}{
		t.name: map[string]interface{}{
			"terms": termsMap,
		},
	}
}
