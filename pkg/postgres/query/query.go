// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

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

// buildComparisonStatementSimple builds a SQL filter statement of the form:
// [NOT] ((field operator ?) OR (field operator ?) OR ...)
// Example: NOT ((field = ?) OR (field = ?)) for an argument with multiple values
func buildComparisonStatementSimple(field filter.RequestField, negate bool, operator string) (string, error) {
	return buildComparisonStatement(field, negate, operator, "?")
}

// buildStringComparisonStatement is same as buildComparisonStatementSimple, but with additional
// validation that the input value(s) are of type string
func buildStringComparisonStatement(field filter.RequestField, negate bool, operator string, parameterStatement string) (string, error) {
	// validate that input only contains string(s)
	if valueList, ok := field.Value.([]any); ok {
		if !ok {
			return "", fmt.Errorf("value is of type %T, expected []any", field.Value)
		}
		for _, element := range valueList {
			if _, ok := element.(string); !ok {
				return "", fmt.Errorf("operator '%s' requires string values, got %T", field.Operator, element)
			}
		}
	} else {
		if _, ok := field.Value.(string); !ok {
			return "", fmt.Errorf("operator '%s' requires a string value", field.Operator)
		}
	}

	return buildComparisonStatement(field, negate, operator, parameterStatement)
}

// buildComparisonStatement builds a SQL filter statement of the form:
// [NOT] ((field operator parameterStatement) OR (field operator parameterStatement))
// [parameterStatement] is a statement involving the `?` parameter placeholder. In the simplest case just `?`.
// Example: NOT ((field ILIKE ? || '%' ))
func buildComparisonStatement(field filter.RequestField, negate bool, operator string, parameterStatement string) (string, error) {
	singleStatement := fmt.Sprintf("%s %s %s", field.Name, operator, parameterStatement)

	var count int
	if valueList, ok := field.Value.([]any); ok {
		count = len(valueList)
	} else {
		count = 1
	}

	return chainStatementsByOr(negate, singleStatement, count), nil
}

func chainStatementsByOr(negate bool, singleStatement string, count int) string {
	builder := strings.Builder{}

	if negate {
		builder.WriteString(" NOT ")
	}
	builder.WriteRune('(')
	for i := range count {
		if i > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteRune('(')
		builder.WriteString(singleStatement)
		builder.WriteRune(')')
	}
	builder.WriteRune(')')

	return builder.String()
}

func sanitizeFilterValue(value any) (sanitizedValue any, err error) {
	if value == nil {
		return filter.RequestField{}, errors.New("field has nil value")
	}

	sanitizedValue = value

	if reflect.TypeOf(value).Kind() == reflect.Slice ||
		reflect.TypeOf(value).Kind() == reflect.Array {
		slice := reflect.ValueOf(value)

		if slice.Len() == 0 { // disallow empty list values, as the there is no clear way to interpret this kind of filter
			return filter.RequestField{}, fmt.Errorf("field has empty list of values")
		}
		// convert to []any, so that handlers don't need to deal with different slice types
		var values []any
		if v, ok := value.([]any); ok {
			values = v
		} else {
			values = make([]any, slice.Len())
			for i := 0; i < slice.Len(); i++ {
				values[i] = slice.Index(i).Interface()
			}
		}
		sanitizedValue = values
	}
	return sanitizedValue, nil
}

// buildDateTruncStatement builds a SQL filter statement of the form:
// [NOT] ((date_trunc('day', field) operator date_trunc('day', ?::timestamp)) OR ...)
// Example: NOT ((date_trunc('day', field) < date_trunc('day', ?::timestamp)) OR ...)
func buildDateTruncStatement(field filter.RequestField, operator string) (string, error) {
	checkType := func(value any) error {
		switch value.(type) {
		case string:
			return nil
		case time.Time:
			return nil
		default:
			return fmt.Errorf("operator '%s' requires a string or time.Time value, got: %T", field.Operator, field.Value)
		}
	}

	// validate that input only contains string(s) or time.Time(s)
	var count int
	if valueList, ok := field.Value.([]any); ok {
		for _, element := range valueList {
			err := checkType(element)
			if err != nil {
				return "", err
			}
		}
		count = len(valueList)
	} else {
		err := checkType(field.Value)
		if err != nil {
			return "", err
		}
		count = 1
	}
	singleStatement := fmt.Sprintf("date_trunc('day', %s) %s date_trunc('day', ?::timestamp)", field.Name, operator)

	return chainStatementsByOr(false, singleStatement, count), nil
}
