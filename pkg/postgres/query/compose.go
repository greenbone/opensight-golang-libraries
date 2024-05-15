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
	conditionParams []any, // Parameters for the SQL condition
	err error, // Error encountered during execution
) {
	// translate filter field to database column name if field mapping exists
	if len(fieldMapping) != 0 {
		dbColumnName, ok := fieldMapping[field.Name]
		if ok {
			field.Name = dbColumnName
		}
	}

	switch field.Operator {
	case filter.CompareOperatorIsEqualTo:
		conditionTemplate, conditionParams, err = simpleOperatorCondition(
			field, valueIsList, "%s =", "%s IN",
		)
	case filter.CompareOperatorIsNotEqualTo:
		conditionTemplate, conditionParams, err = simpleOperatorCondition(
			field, valueIsList, "%s !=", "%s NOT IN",
		)
	default:
		err = errors.Errorf("field '%s' with unknown operator '%s'", field.Name, field.Operator)
	}
	return conditionTemplate, conditionParams, err
}
