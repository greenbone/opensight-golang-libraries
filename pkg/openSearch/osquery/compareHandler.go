// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package osquery

import (
	"fmt"
	"strings"
	"time"

	"github.com/aquasecurity/esquery"
	"github.com/samber/lo"
)

// HandleCompareOperatorIsEqualTo handles is equal to
func HandleCompareOperatorIsEqualTo(fieldName string, fieldValue any) (esquery.Mappable, error) {
	// for list of values
	if values, ok := fieldValue.([]any); ok {
		return esquery.Terms(fieldName, values...), nil
	} else { // for single values
		return esquery.Term(fieldName, fieldValue), nil
	}
}

// HandleCompareOperatorContains handles contains.
// In the index mapping the given field must be a string of type `keyword`.
func HandleCompareOperatorContains(fieldName string, fieldValue any) (esquery.Mappable, error) {
	// for a list of values
	if values, ok := fieldValue.([]any); ok {
		queries := make([]esquery.Mappable, 0, len(values))
		for _, value := range values {

			valueStr, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("operator only supports string values: got value %v of type %T", value, value)
			}
			queries = append(queries, esquery.Wildcard(fieldName, "*"+valueStr+"*"))
		}
		return esquery.Bool().Should(queries...).
			MinimumShouldMatch(1), nil
	} else { // for single values
		value, ok := fieldValue.(string)
		if !ok {
			return nil, fmt.Errorf("operator only supports string values: got value %v of type %T", fieldValue, fieldValue)
		}
		return esquery.Wildcard(fieldName, "*"+value+"*"), nil
	}
}

// HandleCompareOperatorTextContains performs a full text search on the given field.
// For substring matching the index mapping must be a string of type `text` with an `ngram`
// multi-field. Otherwise it will only match whole words.
func HandleCompareOperatorTextContains(fieldName string, fieldValue any) (esquery.Mappable, error) {
	if values, ok := fieldValue.([]any); ok {
		if len(values) == 0 {
			return nil, fmt.Errorf("operator requires at least one value")
		}

		queries := make([]esquery.Mappable, 0, len(values))
		for _, value := range values {
			query, err := textSearchQuery(fieldName, value)
			if err != nil {
				return nil, err
			}
			queries = append(queries, query)
		}

		return esquery.Bool().Should(queries...).MinimumShouldMatch(1), nil
	}

	return textSearchQuery(fieldName, fieldValue)
}

func textSearchQuery(fieldName string, value any) (esquery.Mappable, error) {
	val, ok := value.(string)
	terms := strings.Fields(val)
	if !ok || len(terms) == 0 {
		return nil, fmt.Errorf("operator only supports non-empty string values: got value %v of type %T", value, value)
	}

	query := esquery.Bool().Must( // guard against empty string values after potential analyzers were run
		esquery.Bool().Should(
			esquery.Match(fieldName, value),
			esquery.Match(fieldName+".ngram", value),
		).MinimumShouldMatch(1),
	)
	for _, term := range terms {
		query.Must(esquery.Bool().Should(
			esquery.Match(fieldName, term).
				Operator(esquery.OperatorAnd).
				ZeroTermsQuery(esquery.ZeroTermsAll),
			esquery.Match(fieldName+".ngram", term).
				Operator(esquery.OperatorAnd).
				ZeroTermsQuery(esquery.ZeroTermsAll),
		).MinimumShouldMatch(1))
	}

	return query, nil
}

// HandleCompareOperatorBeginsWith handles begins with
func HandleCompareOperatorBeginsWith(fieldName string, fieldValue any) (esquery.Mappable, error) {
	// for a list of values
	if values, ok := fieldValue.([]any); ok {
		prefixQueries := make([]esquery.Mappable, 0, len(values))
		for _, value := range values {

			valueStr, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("operator only supports string values: got value %v of type %T", value, value)
			}
			prefixQueries = append(prefixQueries, esquery.Prefix(fieldName, valueStr))
		}
		return esquery.Bool().Should(prefixQueries...).
			MinimumShouldMatch(1), nil
	} else { // for single value
		value, ok := fieldValue.(string)
		if !ok {
			return nil, fmt.Errorf("operator only supports string values: got value %v of type %T", fieldValue, fieldValue)
		}
		return esquery.Prefix(fieldName, value), nil
	}
}

// HandleCompareOperatorIsLessThanOrEqualTo handles is less than or equal to
func HandleCompareOperatorIsLessThanOrEqualTo(fieldName string, fieldValue any) (esquery.Mappable, error) {
	if values, ok := fieldValue.([]any); ok {
		return esquery.Bool(). // chain by OR
					Should(
				lo.Map(values, func(value any, _ int) esquery.Mappable {
					return esquery.Range(fieldName).
						Lte(value)
				})...,
			).
			MinimumShouldMatch(1), nil
	} else {
		return esquery.Range(fieldName).
			Lte(fieldValue), nil
	}
}

// HandleCompareOperatorIsGreaterThanOrEqualTo handles is greater than or equal to
func HandleCompareOperatorIsGreaterThanOrEqualTo(fieldName string, fieldValue any) (esquery.Mappable, error) {
	if values, ok := fieldValue.([]any); ok {
		return esquery.Bool(). // chain by OR
					Should(
				lo.Map(values, func(value any, _ int) esquery.Mappable {
					return esquery.Range(fieldName).
						Gte(value)
				})...,
			).
			MinimumShouldMatch(1), nil
	} else {
		return esquery.Range(fieldName).
			Gte(fieldValue), nil
	}
}

// HandleCompareOperatorIsGreaterThan handles is greater than
func HandleCompareOperatorIsGreaterThan(fieldName string, fieldValue any) (esquery.Mappable, error) {
	if values, ok := fieldValue.([]any); ok {
		return esquery.Bool(). // chain by OR
					Should(
				lo.Map(values, func(value any, _ int) esquery.Mappable {
					return esquery.Range(fieldName).
						Gt(value)
				})...,
			).
			MinimumShouldMatch(1), nil
	} else {
		return esquery.Range(fieldName).
			Gt(fieldValue), nil
	}
}

// HandleCompareOperatorIsLessThan handles is less than
func HandleCompareOperatorIsLessThan(fieldName string, fieldValue any) (esquery.Mappable, error) {
	if values, ok := fieldValue.([]any); ok {
		return esquery.Bool(). // chain by OR
					Should(
				lo.Map(values, func(value any, _ int) esquery.Mappable {
					return esquery.Range(fieldName).
						Lt(value)
				})...,
			).
			MinimumShouldMatch(1), nil
	} else {
		return esquery.Range(fieldName).
			Lt(fieldValue), nil
	}
}

func HandleCompareOperatorExists(fieldName string, _ any) (esquery.Mappable, error) {
	return esquery.Exists(fieldName), nil
}

// HandleCompareOperatorBetweenDates constructs an OpenSearch range query for a given date field.
// It accepts a field name and a field value, which must be a slice of exactly 2 elements, representing the start and end of range. Accepted slice types:
// - []time.Time,
// - []string of two RFC3339Nano-formatted strings,
// - []any, containing any combination of time.Time and RFC3339Nano-formatted string.
//
// The generated range query is inclusive of both the lower and upper bounds.
// If a document’s timestamp is exactly equal to the start or end date, it will still match the query.
func HandleCompareOperatorBetweenDates(fieldName string, fieldValue any) (esquery.Mappable, error) {
	validateTimeValue := func(value any) error {
		switch val := value.(type) {
		case time.Time:
		case string:
			_, err := time.Parse(time.RFC3339Nano, val)
			if err != nil {
				return fmt.Errorf("invalid date string format: %w", err)
			}
		default:
			return fmt.Errorf("unsupported type: %T, want: time.Time or string", value)
		}

		return nil
	}

	switch dateValue := fieldValue.(type) {
	case []any:
		if len(dateValue) != 2 {
			return nil, fmt.Errorf("invalid fieldValue length for []any: %v", fieldValue)
		}
		err := validateTimeValue(dateValue[0])
		if err != nil {
			return nil, fmt.Errorf("invalid lower bound: %w", err)
		}
		err = validateTimeValue(dateValue[1])
		if err != nil {
			return nil, fmt.Errorf("invalid upper bound: %w", err)
		}

		return esquery.Range(fieldName).
				Gte(dateValue[0]).
				Lte(dateValue[1]),
			nil
	default:
		return nil, fmt.Errorf("unsupported fieldValue type: %T, want: []string, []time.Time, []any", fieldValue)
	}
}
