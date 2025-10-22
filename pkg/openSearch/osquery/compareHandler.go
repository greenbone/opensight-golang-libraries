// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package osquery

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aquasecurity/esquery"
	"github.com/samber/lo"
)

// HandleCompareOperatorIsEqualTo handles is equal to
func HandleCompareOperatorIsEqualTo(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) (esquery.Mappable, error) {
	return createTermQuery(fieldName, fieldValue, fieldKeys, querySettings)
}

// HandleCompareOperatorContains handles contains.
// In the index mapping the given field must be a string of type `keyword`.
func HandleCompareOperatorContains(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) (esquery.Mappable, error) {
	if values, ok := fieldValue.([]interface{}); ok {
		return esquery.Bool().
			Should(
				lo.Map[interface{}, esquery.Mappable](values, func(value interface{}, _ int) esquery.Mappable {
					return esquery.Wildcard(fieldName, "*"+ValueToString(value)+"*")
				})...,
			).
			MinimumShouldMatch(1), nil
	} else { // for single values
		return esquery.Wildcard(fieldName, "*"+ValueToString(fieldValue)+"*"), nil
	}
}

// HandleCompareOperatorTextContains performs a full text search on the given field.
// In the index mapping it must be a string of type `text`.
func HandleCompareOperatorTextContains(fieldName string, _ []string, fieldValue any, _ *QuerySettings) (esquery.Mappable, error) {
	if values, ok := fieldValue.([]any); ok {
		return esquery.Bool(). // chain by OR
					Should(
				lo.Map(values, func(value any, _ int) esquery.Mappable {
					return esquery.Match(fieldName, ValueToString(value)).MinimumShouldMatch("100%")
				})...,
			).
			MinimumShouldMatch(1), nil
	}
	return esquery.Match(fieldName, ValueToString(fieldValue)).MinimumShouldMatch("100%"), nil // no query term is optional
}

// HandleCompareOperatorBeginsWith handles begins with
func HandleCompareOperatorBeginsWith(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) (esquery.Mappable, error) {
	if values, ok := fieldValue.([]any); ok {
		// for a list of values
		return esquery.Bool().
			Should(
				lo.Map[any, esquery.Mappable](values, func(value any, _ int) esquery.Mappable {
					return esquery.Prefix(fieldName, ValueToString(value))
				})...,
			).
			MinimumShouldMatch(1), nil
	}

	// for single value
	return esquery.Prefix(fieldName, ValueToString(fieldValue)), nil
}

// HandleCompareOperatorNotBeginsWith handles not begins with
func HandleCompareOperatorNotBeginsWith(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) (esquery.Mappable, error) {
	// for list of values
	if values, ok := fieldValue.([]interface{}); ok {
		return esquery.Bool().
			MustNot(
				lo.Map[interface{}, esquery.Mappable](values, func(value interface{}, _ int) esquery.Mappable {
					return esquery.Prefix(fieldName, ValueToString(value))
				})...,
			), nil
	} else { // for single values
		return esquery.Prefix(fieldName, fieldValue.(string)), nil
	}
}

// HandleCompareOperatorIsLessThanOrEqualTo handles is less than or equal to
func HandleCompareOperatorIsLessThanOrEqualTo(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) (esquery.Mappable, error) {
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
func HandleCompareOperatorIsGreaterThanOrEqualTo(fieldName string, fieldKeys []string,
	fieldValue any, querySettings *QuerySettings,
) (esquery.Mappable, error) {
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
func HandleCompareOperatorIsGreaterThan(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) (esquery.Mappable, error) {
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
func HandleCompareOperatorIsLessThan(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) (esquery.Mappable, error) {
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

func createTermQuery(fieldName string, fieldValue any, fieldKeys []string, querySettings *QuerySettings) (esquery.Mappable, error) {
	// for list of values
	if values, ok := fieldValue.([]interface{}); ok {
		return esquery.Terms(fieldName, values...), nil
	} else { // for single values
		return esquery.Term(fieldName, fieldValue), nil
	}
}

func HandleCompareOperatorExists(fieldName string, _ []string, _ any, _ *QuerySettings) (esquery.Mappable, error) {
	return esquery.Exists(fieldName), nil
}

// ValueToString converts the given value to a string.
// Compared to [fmt.Sprint] it will give RFC3339 format for [time.Time] value
// and a specific formatting of numbers.
func ValueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprint(value)
	}
}

// HandleCompareOperatorBetweenDates constructs an OpenSearch range query for a given date field.
// It accepts a field name and a field value, which must be a slice of exactly 2 elements, representing the start and end of range. Accepted slice types:
// - []time.Time,
// - []string of two RFC3339Nano-formatted strings,
// - []any, containing any combination of time.Time and RFC3339Nano-formatted string.
//
// The generated range query is inclusive of both the lower and upper bounds.
// If a document’s timestamp is exactly equal to the start or end date, it will still match the query.
func HandleCompareOperatorBetweenDates(fieldName string, _ []string, fieldValue any, _ *QuerySettings) (esquery.Mappable, error) {
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
