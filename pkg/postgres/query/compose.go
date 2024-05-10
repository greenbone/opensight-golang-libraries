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
	valueList []any, // List of values for comparison
) (
	conditionTemplate string, // Template for the SQL condition
	conditionParams []any, // Parameters for the SQL condition
	err error, // Error encountered during execution
) {
	switch field.Operator {
	case filter.CompareOperatorIsEqualTo:
		_, ok := fieldMapping[field.Name]
		if ok {
			conditionTemplate, conditionParams, err = simpleOperatorCondition(
				field, valueIsList, "%s =", "%s IN",
			)
		} else {
			conditionTemplate, conditionParams, err = likeOperatorCondition(
				field, valueIsList, valueList, equalsAsLikePattern, false,
			)
		}
	case filter.CompareOperatorIsNotEqualTo:
		_, ok := fieldMapping[field.Name]
		if ok {
			conditionTemplate, conditionParams, err = simpleOperatorCondition(
				field, valueIsList, "%s !=", "%s NOT IN",
			)
		} else {
			conditionTemplate, conditionParams, err = likeOperatorCondition(
				field, valueIsList, valueList, equalsAsLikePattern, true,
			)
		}
	case filter.CompareOperatorIsLessThan:
		conditionTemplate, conditionParams, err = simpleOperatorCondition(
			field, valueIsList, "%s <", "%s < greatest",
		)
	case filter.CompareOperatorIsLessThanOrEqualTo:
		conditionTemplate, conditionParams, err = simpleOperatorCondition(
			field, valueIsList, "%s <=", "%s <= greatest",
		)
	case filter.CompareOperatorIsGreaterThan:
		conditionTemplate, conditionParams, err = simpleOperatorCondition(
			field, valueIsList, "%s >", "%s > least",
		)
	case filter.CompareOperatorIsGreaterThanOrEqualTo:
		conditionTemplate, conditionParams, err = simpleOperatorCondition(
			field, valueIsList, "%s >=", "%s >= least",
		)
	case filter.CompareOperatorContains:
		conditionTemplate, conditionParams, err = likeOperatorCondition(
			field, valueIsList, valueList, containsAsLikePattern, false,
		)
	case filter.CompareOperatorDoesNotContain:
		conditionTemplate, conditionParams, err = likeOperatorCondition(
			field, valueIsList, valueList, containsAsLikePattern, true,
		)
	case filter.CompareOperatorBeginsWith:
		conditionTemplate, conditionParams, err = likeOperatorCondition(
			field, valueIsList, valueList, beginsWithAsLikePattern, false,
		)
	case filter.CompareOperatorDoesNotBeginWith:
		conditionTemplate, conditionParams, err = likeOperatorCondition(
			field, valueIsList, valueList, beginsWithAsLikePattern, true,
		)
	case filter.CompareOperatorBeforeDate:
		conditionTemplate, conditionParams, err = simpleSingleStringValueOperatorCondition(
			field, valueIsList,
			"date_trunc('day'::text, %s) < date_trunc('day'::text, ::timestamp)",
		)
	case filter.CompareOperatorAfterDate:
		conditionTemplate, conditionParams, err = simpleSingleStringValueOperatorCondition(
			field, valueIsList,
			"date_trunc('day'::text, %s) > date_trunc('day'::text, ::timestamp)",
		)
	default:
		err = errors.Errorf("field '%s' with unknown operator '%s'", field.Name, field.Operator)
	}
	return conditionTemplate, conditionParams, err
}
