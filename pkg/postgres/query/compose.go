// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/pkg/errors"
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
			"Mapping for filter field '%s' is currently not implemented.", field.Name)
	}

	switch field.Operator {
	case filter.CompareOperatorIsEqualTo:
		conditionTemplate, err = simpleOperatorCondition(
			field, valueIsList, " %s = ?", " %s IN (%s)",
		)
	case filter.CompareOperatorIsNotEqualTo:
		conditionTemplate, err = simpleOperatorCondition(
			field, valueIsList, " %s != ?", " %s NOT IN (%s)",
		)
	case filter.CompareOperatorIsLessThan:
		conditionTemplate, err = simpleOperatorCondition(
			field, valueIsList, " %s < ?", " %s < LEAST(%s)",
		)
	case filter.CompareOperatorIsLessThanOrEqualTo:
		conditionTemplate, err = simpleOperatorCondition(
			field, valueIsList, " %s <= ?", " %s <= LEAST(%s)",
		)
	case filter.CompareOperatorIsGreaterThan:
		conditionTemplate, err = simpleOperatorCondition(
			field, valueIsList, " %s > ?", " %s > GREATEST(%s)",
		)
	case filter.CompareOperatorIsGreaterThanOrEqualTo:
		conditionTemplate, err = simpleOperatorCondition(
			field, valueIsList, " %s >= ?", " %s >= GREATEST(%s)",
		)
	case filter.CompareOperatorContains:
		conditionTemplate, err = likeOperatorCondition(
			field, valueIsList, false, false,
		)
	case filter.CompareOperatorBeginsWith:
		conditionTemplate, err = likeOperatorCondition(
			field, valueIsList, false, true,
		)
	case filter.CompareOperatorBeforeDate:
		conditionTemplate, err = simpleSingleStringValueOperatorCondition(
			field, valueIsList,
			" date_trunc('day'::text, %s) < date_trunc('day'::text, ?::timestamp)",
		)
	case filter.CompareOperatorAfterDate:
		conditionTemplate, err = simpleSingleStringValueOperatorCondition(
			field, valueIsList,
			" date_trunc('day'::text, %s) > date_trunc('day'::text, ?::timestamp)",
		)
	default:
		err = errors.Errorf("field '%s' with unknown operator '%s'", field.Name, field.Operator)
	}
	return conditionTemplate, err
}
