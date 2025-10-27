// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/lib/pq"
)

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
) (conditionTemplate string, err error) {
	quotedName, err := getQuotedName(field.Name)
	if err != nil {
		return "", fmt.Errorf("failed to parse quoted name: %w", err)
	}

	if valueIsList {
		valueList, ok := field.Value.([]any)
		if !ok {
			return "", fmt.Errorf("value is of type %T, expected []any", field.Value)
		}
		placeholders := make([]string, len(valueList))
		for i := range valueList {
			placeholders[i] = "?"
		}
		conditionTemplate = fmt.Sprintf(listValueTemplate, quotedName, strings.Join(placeholders, ", "))
	} else {
		conditionTemplate = fmt.Sprintf(singleValueTemplate, quotedName)
	}

	return conditionTemplate, nil
}

func checkFieldValueType(field filter.RequestField) (valueIsList bool, valueList []any, err error) {
	if reflect.TypeOf(field.Value).Kind() == reflect.Slice {
		valueList, valueIsList = field.Value.([]any)
		if !valueIsList {
			err = fmt.Errorf("list field '%s' must have type []any, got %T", field.Name, field.Value)
			return false, nil, err
		}
	} else {
		valueIsList = false
		valueList = nil
	}
	return valueIsList, valueList, nil
}

// likeOperatorCondition generates a SQL 'LIKE' condition template based on the given field parameters.
// - negate: a boolean indicating if the condition should be negated (e.g., NOT LIKE).
// - beginsWith: a boolean indicating if the condition should check for "begins with" and use beginsWithAsLikePattern().
// if beginsWith is false, the condition should check for "contains" and use containsAsLikePattern()
func likeOperatorCondition(
	field filter.RequestField, valueIsList bool, negate bool, beginsWith bool,
) (conditionTemplate string, err error) {
	quotedName, err := getQuotedName(field.Name)
	if err != nil {
		return "", fmt.Errorf("could not get quoted name: %w", err)
	}

	if valueIsList {
		valueList, ok := field.Value.([]any)
		if !ok {
			return "", fmt.Errorf("value is of type %T, expected []any", field.Value)
		}
		for _, element := range valueList {
			if _, ok := element.(string); !ok {
				err = fmt.Errorf(
					"operator '%s' requires string values, got %T",
					field.Operator, element,
				)
				return "", err
			}
		}

		conditionTemplate = multiLikeOrTemplate(quotedName, len(valueList), negate, beginsWith)
	} else {
		if _, ok := field.Value.(string); ok {
			conditionTemplate = singleLikeTemplate(quotedName, negate, beginsWith)
		} else {
			err = fmt.Errorf("operator '%s' requires a string value", field.Operator)
			return "", err
		}
	}

	return conditionTemplate, nil
}

func singleLikeTemplate(quotedField string, negate bool, beginsWith bool) string {
	if negate {
		return quotedField + " NOT ILIKE " + handleMultiLikeType(beginsWith)
	} else {
		return quotedField + " ILIKE " + handleMultiLikeType(beginsWith)
	}
}

func multiLikeOrTemplate(quotedField string, elementCount int, negate bool, beginsWith bool) string {
	builder := strings.Builder{}
	if negate {
		builder.WriteString("NOT ")
	}
	builder.WriteString(" (")
	for i := 0; i < elementCount; i++ {
		if i > 0 {
			builder.WriteString(") OR (")
		}
		builder.WriteString(quotedField + " ILIKE " + handleMultiLikeType(beginsWith))
	}
	builder.WriteRune(')')
	return builder.String()
}

func handleMultiLikeType(beginsWith bool) string {
	if beginsWith {
		return beginsWithAsLikePattern()
	}
	return containsAsLikePattern()
}

func containsAsLikePattern() string {
	return `'%' || ? || '%'`
}

func beginsWithAsLikePattern() string {
	return `? || '%'`
}

func simpleSingleStringValueOperatorCondition(
	field filter.RequestField, valueIsList bool, singleValueTemplate string,
) (conditionTemplate string, err error) {
	if valueIsList {
		err = fmt.Errorf("operator '%s' does not support multi-select", field.Operator)
		return "", err
	} else if _, ok := field.Value.(string); ok {
		quotedName, err := getQuotedName(field.Name)
		if err != nil {
			return "", fmt.Errorf("could not get quoted name: %w", err)
		}
		conditionTemplate = fmt.Sprintf(singleValueTemplate, quotedName)
	} else {
		err = fmt.Errorf("operator '%s' requires a string value", field.Operator)
		return "", err
	}
	return conditionTemplate, nil
}
