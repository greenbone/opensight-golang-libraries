// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/lib/pq"
	"github.com/pkg/errors"
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