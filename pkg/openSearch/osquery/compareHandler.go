// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package osquery

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/aquasecurity/esquery"
	"github.com/samber/lo"
)

// HandleCompareOperatorIsEqualTo handles is equal to
func HandleCompareOperatorIsEqualTo(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return createTermQuery(fieldName, fieldValue, fieldKeys, querySettings)
}

// HandleCompareOperatorContains handles contains.
// In the index mapping the given field must be a string of type `keyword`.
func HandleCompareOperatorContains(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	if values, ok := fieldValue.([]interface{}); ok {
		return esquery.Bool().
			Should(
				lo.Map[interface{}, esquery.Mappable](values, func(value interface{}, _ int) esquery.Mappable {
					return esquery.Wildcard(fieldName, "*"+ValueToString(value)+"*")
				})...,
			).
			MinimumShouldMatch(1)
	} else { // for single values
		return esquery.Wildcard(fieldName, "*"+ValueToString(fieldValue)+"*")
	}
}

// HandleCompareOperatorTextContains performs a full text search on the given field.
// In the index mapping it must be a string of type `text`.
func HandleCompareOperatorTextContains(fieldName string, _ []string, fieldValue any, _ *QuerySettings) esquery.Mappable {
	var value string
	if values, ok := fieldValue.([]any); ok { // value as list and value as space separated string should result in same query
		stringValues := make([]string, 0, len(values))
		for _, val := range values {
			stringValues = append(stringValues, ValueToString(val))
		}
		value = strings.Join(stringValues, " ")
	} else {
		value = ValueToString(fieldValue)
	}
	return esquery.Match(fieldName, value).MinimumShouldMatch("100%") // no query term is optional
}

// HandleCompareOperatorBeginsWith handles begins with
func HandleCompareOperatorBeginsWith(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return handleCompareOperatorBeginsWith(fieldName, fieldValue)
}

func handleCompareOperatorBeginsWith(fieldName string, fieldValue any) esquery.Mappable {
	if values, ok := fieldValue.([]any); ok {
		// for a list of values
		return esquery.Bool().
			Should(
				lo.Map[any, esquery.Mappable](values, func(value any, _ int) esquery.Mappable {
					return esquery.Prefix(fieldName, ValueToString(value))
				})...,
			).
			MinimumShouldMatch(1)
	}

	// for single value
	return esquery.Prefix(fieldName, ValueToString(fieldValue))
}

// HandleCompareOperatorNotBeginsWith handles not begins with
func HandleCompareOperatorNotBeginsWith(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	// for list of values
	if values, ok := fieldValue.([]interface{}); ok {
		return esquery.Bool().
			MustNot(
				lo.Map[interface{}, esquery.Mappable](values, func(value interface{}, _ int) esquery.Mappable {
					return esquery.Prefix(fieldName, ValueToString(value))
				})...,
			)
	} else { // for single values
		return esquery.Prefix(fieldName, fieldValue.(string))
	}
}

// HandleCompareOperatorIsLessThanOrEqualTo handles is less than or equal to
func HandleCompareOperatorIsLessThanOrEqualTo(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return esquery.Range(fieldName).
		Lte(fieldValue)
}

// HandleCompareOperatorIsGreaterThanOrEqualTo handles is greater than or equal to
func HandleCompareOperatorIsGreaterThanOrEqualTo(fieldName string, fieldKeys []string,
	fieldValue any, querySettings *QuerySettings,
) esquery.Mappable {
	return esquery.Range(fieldName).
		Gte(fieldValue)
}

// HandleCompareOperatorIsGreaterThan handles is greater than
func HandleCompareOperatorIsGreaterThan(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return esquery.Range(fieldName).
		Gt(fieldValue)
}

// HandleCompareOperatorIsLessThan handles is less than
func HandleCompareOperatorIsLessThan(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return esquery.Range(fieldName).
		Lt(fieldValue)
}

func createTermQuery(fieldName string, fieldValue any, fieldKeys []string, querySettings *QuerySettings) esquery.Mappable {
	// for list of values
	if values, ok := fieldValue.([]interface{}); ok {
		return esquery.Terms(fieldName, values...)
	} else { // for single values
		return esquery.Term(fieldName, fieldValue)
	}
}

func HandleCompareOperatorExists(fieldName string, _ []string, _ any, _ *QuerySettings) esquery.Mappable {
	return esquery.Exists(fieldName)
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
// It accepts a field name and a field value, which must be either:
// - A slice of two time.Time values ([]time.Time), representing the start and end of the range, or
// - A slice of two RFC3339Nano-formatted strings ([]string), which are parsed into time.Time,  representing the start and end of the range.
//
// The generated range query is inclusive of both the lower and upper bounds.
// If a document’s timestamp is exactly equal to the start or end date, it will still match the query.
//
// If the slice length is not exactly 2, or if the string values cannot be parsed into valid dates,
// the function logs an error and returns an empty query (MatchNone).
func HandleCompareOperatorBetweenDates(fieldName string, _ []string, fieldValue any, _ *QuerySettings) esquery.Mappable {
	switch dateValue := fieldValue.(type) {
	case []time.Time:
		if len(dateValue) != 2 {
			log.Error().Msgf("invalid fieldValue length for []time.Time: %T", fieldValue)
			return esquery.MatchNone()
		}
		return esquery.Range(fieldName).
			Gte(dateValue[0]).
			Lte(dateValue[1])
	case []string:
		if len(dateValue) != 2 {
			log.Error().Msgf("invalid fieldValue length for []string: %T", fieldValue)
			return esquery.MatchNone()
		}
		start, err1 := time.Parse(time.RFC3339Nano, dateValue[0])
		end, err2 := time.Parse(time.RFC3339Nano, dateValue[1])
		if err1 != nil || err2 != nil {
			log.Error().Msgf("invalid date format in []string: %v, %v", err1, err2)
			return esquery.MatchNone()
		}
		return esquery.Range(fieldName).
			Gte(start).
			Lte(end)
	default:
		log.Error().Msgf("unsupported fieldValue type: %T, want: []string, []time.Time", fieldValue)
		return esquery.MatchNone()
	}
}
