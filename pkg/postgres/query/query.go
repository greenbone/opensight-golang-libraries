// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"fmt"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

// Replacer for escaping LIKE clause wildcards and backslashes
var likeReplacer = strings.NewReplacer(`_`, `\_`, `%`, `\%`, `\`, `\\`)

func getQuotedName(fieldName string) (string, error) {
	// handles field names for joins that contain a `.`
	if strings.Contains(fieldName, ".") {
		split := strings.SplitN(fieldName, ".", 2)
		if len(split[0]) < 1 {
			return "", errors.New("table name can not be empty")
		}
		if split[1] == "" {
			return "", errors.New("field name can not be empty")
		}
		tableName := split[0]
		fieldName := split[1]
		return pq.QuoteIdentifier(tableName) + "." + pq.QuoteIdentifier(fieldName), nil

	}

	return pq.QuoteIdentifier(fieldName), nil
}

func simpleOperatorCondition(
	field filter.RequestField, valueIsList bool, singleValueTemplate string, listValueTemplate string,
) (conditionTemplate string, conditionParams []any, err error) {
	quotedName, err := getQuotedName(field.Name)
	if err != nil {
		return "", nil, errors.Wrap(err, "could not get quoted name")
	}

	conditionParams = []any{field.Value}
	if valueIsList {
		conditionTemplate = fmt.Sprintf(listValueTemplate, quotedName)
	} else {
		conditionTemplate = fmt.Sprintf(singleValueTemplate, quotedName)
	}
	return conditionTemplate, conditionParams, nil
}

func singleLikeTemplate(quotedField string, negate bool) string {
	if negate {
		return quotedField + " NOT ILIKE ?"
	} else {
		return quotedField + " ILIKE ?"
	}
}

func equalsAsLikePattern(value string) string {
	return likeReplacer.Replace(value)
}

func containsAsLikePattern(value string) string {
	return "%" + likeReplacer.Replace(value) + "%"
}

func beginsWithAsLikePattern(value string) string {
	return likeReplacer.Replace(value) + "%"
}

func multiLikeOrTemplate(quotedField string, elementCount int, negate bool) string {
	builder := strings.Builder{}
	if negate {
		builder.WriteString("NOT ")
	}
	builder.WriteRune('(')
	for i := 0; i < elementCount; i++ {
		if i > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString(quotedField + " ILIKE ?")
	}
	builder.WriteRune(')')
	return builder.String()
}

func likeOperatorCondition(
	field filter.RequestField, valueIsList bool, valueList []any, likeValueFunc func(string) string, negate bool,
) (conditionTemplate string, conditionParams []any, err error) {
	quotedName, err := getQuotedName(field.Name)
	if err != nil {
		return "", nil, errors.Wrap(err, "could not get quoted name")
	}

	if valueIsList {
		conditionParams = make([]any, len(valueList))
		for i, element := range valueList {
			if elementStr, ok := element.(string); ok {
				conditionParams[i] = likeValueFunc(elementStr)
				conditionTemplate = multiLikeOrTemplate(quotedName, len(valueList), negate)
			} else {
				err = errors.Errorf(
					"operator '%s' requires string values, got %T",
					field.Operator, element,
				)
				return "", nil, err
			}
		}
	} else {
		if valueStr, ok := field.Value.(string); ok {
			conditionParams = []any{likeValueFunc(valueStr)}
			conditionTemplate = singleLikeTemplate(quotedName, negate)
		} else {
			err = errors.Errorf("operator '%s' requires a string value", field.Operator)
			return "", nil, err
		}
	}

	return conditionTemplate, conditionParams, nil
}

func simpleSingleStringValueOperatorCondition(
	field filter.RequestField, valueIsList bool, singleValueTemplate string,
) (conditionTemplate string, conditionParams []any, err error) {
	conditionParams = []any{field.Value}
	if valueIsList {
		err = errors.Errorf("operator '%s' does not support multi-select", field.Operator)
		return "", nil, err
	} else if _, ok := field.Value.(string); ok {
		quotedName, err := getQuotedName(field.Name)
		if err != nil {
			return "", nil, errors.Wrap(err, "could not get quoted name")
		}
		conditionTemplate = fmt.Sprintf(singleValueTemplate, quotedName)
	} else {
		err = errors.Errorf("operator '%s' requires a string value", field.Operator)
		return "", nil, err
	}
	return conditionTemplate, conditionParams, nil
}

func checkFieldValueType(field filter.RequestField) (valueIsList bool, valueList []any, err error) {
	if reflect.TypeOf(field.Value).Kind() == reflect.Slice {
		valueList, valueIsList = field.Value.([]any)
		if !valueIsList {
			err = errors.Errorf("list field '%s' must have type []any, got %T", field.Name, field.Value)
			return false, nil, err
		}
	} else {
		valueIsList = false
		valueList = nil
	}
	return valueIsList, valueList, nil
}
