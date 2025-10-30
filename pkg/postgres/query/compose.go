// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"fmt"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
)

// composeQuery takes a field mapping, a filter request field, a flag indicating whether the value is a list,
// and a list of values. It returns a condition template, a list of condition parameters, and an error.
func composeQuery(
	fieldMapping map[string]string, // Mapping of field names to database column names
	field filter.RequestField, // The filter request field containing the field name and operator
	valueIsList bool, // Indicates if the value is a list
) (
	conditionTemplate string, // Template for the SQL condition
	err error, // Error encountered during execution
) {
	// translate filter field to database column name if field mapping exists
	dbColumnName, ok := fieldMapping[field.Name]
	if ok {
		field.Name = dbColumnName
	} else {
		return "", filter.NewInvalidFilterFieldError(
			"mapping for filter field '%s' is currently not implemented", field.Name)
	}

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
