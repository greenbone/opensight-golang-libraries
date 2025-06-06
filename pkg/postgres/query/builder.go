// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package query facilitates the translation of a result selector into a PostgresSQL conditional query string, incorporating sorting and paging functionalities.
package query

import (
	"errors"
	"fmt"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/query"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
)

// Settings is a configuration struct used to customize the behavior of the query builder.
type Settings struct {
	// FilterFieldMapping is the mapping of filter fields for query customization
	// also serves as safeguard against sql injection
	FilterFieldMapping map[string]string
}

// Builder represents a query builder used to construct PostgresSQL conditional query strings
// with sorting and paging functionalities.
type Builder struct {
	querySettings *Settings       // Settings used to configure the query builder
	query         strings.Builder // strings.Builder to construct the query string
}

// NewPostgresQueryBuilder creates a new instance of the query builder with the provided settings.
func NewPostgresQueryBuilder(querySetting *Settings) *Builder {
	return &Builder{
		querySettings: querySetting,
	}
}

// BuildQueryConditions builds and appends filter conditions to the query builder based on the provided filter request.
// It constructs conditional clauses using the logic operator specified in the request.
// It uses the `?` query placeholder, so you can pass your parameter separately
// It returns all individual field values in a single list
// BuildQueryConditions can be used as a standalone function with Gorm
func (qb *Builder) BuildQueryConditions(request *filter.Request) (args []any, err error) {
	if request == nil || len(request.Fields) == 0 {
		return nil, nil
	}
	logicOperator := strings.ToUpper(string(request.Operator))

	qb.query.WriteString("WHERE")

	for index, field := range request.Fields {
		var (
			err               error
			valueIsList       bool
			conditionTemplate string
		)
		valueIsList, _, err = checkFieldValueType(field)
		if err != nil {
			return nil, fmt.Errorf("error checking filter field value type '%s': %w", field, err)
		}
		conditionTemplate, err = composeQuery(qb.querySettings.FilterFieldMapping, field, valueIsList)
		if err != nil {
			return nil, fmt.Errorf("error composing query from filter field %w", err)
		}
		if index > 0 {
			qb.query.WriteString(fmt.Sprintf(" %s", logicOperator))
		}
		args = append(args, extractFieldValues(field.Value, field.Operator)...)
		qb.query.WriteString(conditionTemplate)
	}
	return
}

// likeReplacer is used for escaping LIKE and ILIKE clauses wildcards and backslashes
var likeReplacer = strings.NewReplacer(`_`, `\_`, `%`, `\%`, `\`, `\\`)

func extractFieldValues(input any, compareOperator filter.CompareOperator) (resp []any) {
	processString := func(str string) string {
		switch compareOperator {
		case
			// compareOperators using LIKE or ILIKE operators
			filter.CompareOperatorBeginsWith,
			filter.CompareOperatorContains,
			filter.CompareOperatorIsStringCaseInsensitiveEqualTo:

			return likeReplacer.Replace(str)

		default:
			return str
		}
	}

	if values, isSlice := input.([]any); isSlice {
		// validate values and escape special symbols
		for index, value := range values {
			if strValue, isString := value.(string); isString {
				values[index] = processString(strValue)
			}
		}
		return values
	}

	if strValue, isString := input.(string); isString {
		return []any{processString(strValue)}
	}
	return []any{input}
}

// addSorting appends sorting conditions to the query builder based on the provided sorting request.
// It constructs the ORDER BY clause using the specified sort column and direction.
func (qb *Builder) addSorting(sort *sorting.Request) error {
	if sort == nil {
		return errors.New("missing sorting fields, add sort request or remove call to addSorting()")
	}

	dbColumnName, ok := qb.querySettings.FilterFieldMapping[sort.SortColumn]
	if !ok {
		return filter.NewInvalidFilterFieldError(
			"missing filter field mapping for '%s'", sort.SortColumn)
	}

	qb.query.WriteString(fmt.Sprintf(" ORDER BY %s %s", dbColumnName, sort.SortDirection))
	return nil
}

// addPaging appends paging conditions to the query builder based on the provided paging request.
// It constructs the OFFSET and LIMIT clauses according to the specified page index and page size.
func (qb *Builder) addPaging(paging *paging.Request) error {
	if paging == nil {
		return errors.New("missing paging fields, add paging request or remove call to AddSize()")
	}

	if paging.PageIndex > 0 {
		offset := paging.PageIndex * paging.PageSize
		qb.query.WriteString(fmt.Sprintf(" OFFSET %d", offset))
	}

	if paging.PageSize > 0 {
		qb.query.WriteString(fmt.Sprintf(" LIMIT %d", paging.PageSize))
	}
	return nil
}

// Build generates the complete SQL query based on the provided result selector.
// It constructs the query by adding filter, sorting, and paging conditions.
// It returns the constructed query string, and all the individual filter fields values (args) in a single list
// If any error occurs during the construction, it returns an empty string.
func (qb *Builder) Build(resultSelector query.ResultSelector) (query string, args []any, err error) {
	if resultSelector.Filter != nil {
		args, err = qb.BuildQueryConditions(resultSelector.Filter)
		if err != nil {
			err = fmt.Errorf("error adding filter query: %w", err)
			return
		}
	}

	if resultSelector.Sorting != nil {
		sortingErr := qb.addSorting(resultSelector.Sorting)
		if sortingErr != nil {
			err = fmt.Errorf("error adding sort query: %w", sortingErr)
			return
		}
	}

	if resultSelector.Paging != nil {
		err = qb.addPaging(resultSelector.Paging)
		if err != nil {
			err = fmt.Errorf("error adding paging query: %w", err)
			return
		}
	}

	query = qb.query.String()
	return
}
