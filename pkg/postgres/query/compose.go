// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"fmt"
	"maps"
	"slices"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
)

// composeQuery takes a filter request field and translates it into a SQL query condition
// which can be used in a WHERE clause.
func composeQuery(
	fieldMapping map[string]string, // Mapping of field names to database column names
	field filter.RequestField, // The filter request field containing the field name and operator
) (
	conditionTemplate string, // Template for the SQL condition
	err error, // Error encountered during execution
) {
	// translate filter field to database column name if field mapping exists
	dbColumnName, ok := fieldMapping[field.Name]
	if !ok {
		return "", filter.NewInvalidFilterFieldError(
			"invalid filter field '%s', available fields: ",
			slices.Collect(maps.Keys(fieldMapping)))
	}
	quotedName, err := getQuotedName(dbColumnName)
	if err != nil {
		return "", fmt.Errorf("failed to parse quoted name of field %s: %w", field.Name, err)
	}
	field.Name = quotedName

	switch field.Operator {
	case filter.CompareOperatorIsEqualTo:
		conditionTemplate, err = buildComparisonStatementSimple(field, false, "=")
	case filter.CompareOperatorIsNotEqualTo:
		conditionTemplate, err = buildComparisonStatementSimple(field, true, "=")
	case filter.CompareOperatorIsLessThan:
		conditionTemplate, err = buildComparisonStatementSimple(field, false, "<")
	case filter.CompareOperatorIsLessThanOrEqualTo:
		conditionTemplate, err = buildComparisonStatementSimple(field, false, "<=")
	case filter.CompareOperatorIsGreaterThan:
		conditionTemplate, err = buildComparisonStatementSimple(field, false, ">")
	case filter.CompareOperatorIsGreaterThanOrEqualTo:
		conditionTemplate, err = buildComparisonStatementSimple(field, false, ">=")
	case filter.CompareOperatorContains:
		conditionTemplate, err = buildStringComparisonStatement(field, false, "ILIKE", `'%' || ? || '%'`)
	case filter.CompareOperatorDoesNotContain:
		conditionTemplate, err = buildStringComparisonStatement(field, true, "ILIKE", `'%' || ? || '%'`)
	case filter.CompareOperatorBeginsWith:
		conditionTemplate, err = buildStringComparisonStatement(field, false, "ILIKE", `? || '%'`)
	case filter.CompareOperatorDoesNotBeginWith:
		conditionTemplate, err = buildStringComparisonStatement(field, true, "ILIKE", `? || '%'`)
	case filter.CompareOperatorIsStringCaseInsensitiveEqualTo:
		conditionTemplate, err = buildStringComparisonStatement(field, false, "ILIKE", "?")
	case filter.CompareOperatorBeforeDate:
		conditionTemplate, err = buildDateTruncStatement(field, "<")
	case filter.CompareOperatorAfterDate:
		conditionTemplate, err = buildDateTruncStatement(field, ">")
	default:
		err = fmt.Errorf("field '%s' with unknown operator '%s'", field.Name, field.Operator)
	}
	return conditionTemplate, err
}
