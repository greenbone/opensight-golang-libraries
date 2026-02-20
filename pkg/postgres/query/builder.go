// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package query facilitates the translation of a result selector into a PostgresSQL conditional query string, incorporating sorting and paging functionalities.
package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/query"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
)

// Settings is a configuration struct used to customize the behavior of the query builder.
type Settings struct {
	// FilterFieldMapping is the mapping of filter (or sorting) fields to columns in the database.
	// The columns will be part of the `WHERE` or `ORDER BY` clause, so depending on the query the entry
	// needs to be prefixed with the table name or alias used in the query.
	// It also serves as safeguard against sql injection.
	FilterFieldMapping map[string]string
	// Column used as tie breaker when sorting results. It should have a unique value for each row.
	// It will be part of the `ORDER BY` clause, so depending on the query the entry needs to be prefixed
	// with the table name or alias used in the query.
	SortingTieBreakerColumn string
}

// Builder represents a query builder used to construct PostgresSQL conditional query strings
// with sorting and paging functionalities.
type Builder struct {
	querySettings Settings        // Settings used to configure the query builder
	query         strings.Builder // strings.Builder to construct the query string
}

// NewPostgresQueryBuilder creates a new instance of the query builder with the provided settings.
func NewPostgresQueryBuilder(querySetting Settings) (*Builder, error) {
	if querySetting.SortingTieBreakerColumn == "" {
		return nil, fmt.Errorf("missing sorting tie breaker column in query settings")
	}

	return &Builder{querySettings: querySetting}, nil
}

// addFilters builds and appends filter conditions to the query builder based on the provided filter request.
// It constructs conditional clauses using the logic operator specified in the request.
// It uses the `?` query placeholder, so you can pass your parameter separately
// It returns all individual field values in a single list
func (qb *Builder) addFilters(request *filter.Request) (args []any, err error) {
	if request == nil || len(request.Fields) == 0 {
		return nil, nil
	}
	if request.Operator == "" && len(request.Fields) == 1 { // for single filter `Operator` is not relevant
		request.Operator = filter.LogicOperatorAnd
	}
	var logicOperator string
	switch request.Operator {
	case filter.LogicOperatorAnd:
		logicOperator = "AND"
	case filter.LogicOperatorOr:
		logicOperator = "OR"
	default:
		return nil, fmt.Errorf("invalid filter logic operator: %s", request.Operator)
	}

	qb.query.WriteString("WHERE ")
	for index, field := range request.Fields {
		sanitizedValue, err := sanitizeFilterValue(field.Value)
		if err != nil {
			return nil, fmt.Errorf("error sanitizing filter field value '%s': %w", field.Name, err)
		}
		field.Value = sanitizedValue
		args = append(args, extractFieldValues(field.Value, field.Operator)...)

		conditionTemplate, err := composeQuery(qb.querySettings.FilterFieldMapping, field)
		if err != nil {
			return nil, fmt.Errorf("error composing query from filter field %q:  %w", field.Name, err)
		}

		if index > 0 {
			fmt.Fprintf(&qb.query, " %s ", logicOperator)
		}
		qb.query.WriteString(conditionTemplate)
	}
	return args, nil
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
	sortStatement := " ORDER BY"

	if sort != nil {
		var sortDirection string
		switch sort.SortDirection {
		case sorting.DirectionAscending:
			sortDirection = "ASC"
		case sorting.DirectionDescending:
			sortDirection = "DESC"
		default:
			return fmt.Errorf("invalid sort direction: %s", sort.SortDirection)
		}

		dbColumnName, ok := qb.querySettings.FilterFieldMapping[sort.SortColumn]
		if !ok {
			return filter.NewInvalidFilterFieldError(
				"missing filter field mapping for '%s'", sort.SortColumn)
		}
		sortStatement += fmt.Sprintf(" %s %s,", dbColumnName, sortDirection)
	}
	// add tie breaker to ensure consistent sorting
	sortStatement += fmt.Sprintf(" %s ASC", qb.querySettings.SortingTieBreakerColumn)

	qb.query.WriteString(sortStatement)
	return nil
}

// addPaging appends paging conditions to the query builder based on the provided paging request.
// It constructs the OFFSET and LIMIT clauses according to the specified page index and page size.
func (qb *Builder) addPaging(paging paging.Request) error {
	if paging.PageSize < 0 || paging.PageIndex < 0 {
		return fmt.Errorf("paging parameters must be non-negative, got page size: %d, page index: %d",
			paging.PageSize, paging.PageIndex)
	}

	if paging.PageIndex > 0 {
		offset := paging.PageIndex * paging.PageSize
		fmt.Fprintf(&qb.query, " OFFSET %d", offset)
	}

	if paging.PageSize > 0 {
		fmt.Fprintf(&qb.query, " LIMIT %d", paging.PageSize)
	}
	return nil
}

// Build generates the complete postgres SQL query based on the provided result selector.
// It constructs the query by adding filter, sorting, and paging conditions.
// It returns the constructed query string, and all the individual filter fields values (args) in a single list
func (qb *Builder) Build(resultSelector query.ResultSelector) (query string, args []any, err error) {
	if resultSelector.Filter != nil {
		args, err = qb.addFilters(resultSelector.Filter)
		if err != nil {
			return "", nil, fmt.Errorf("error adding filter query: %w", err)
		}
	}

	err = qb.addSorting(resultSelector.Sorting) // sorting is always applied
	if err != nil {
		return "", nil, fmt.Errorf("error adding sort query: %w", err)
	}

	if resultSelector.Paging != nil {
		err = qb.addPaging(*resultSelector.Paging)
		if err != nil {
			return "", nil, fmt.Errorf("error adding paging query: %w", err)
		}
	}

	query = qb.query.String()
	query = rebind(query)
	return query, args, nil
}

// rebind replaces `?` placeholders with `$1`, `$2`, ... for Postgres compatibility.
// Taken from https://github.com/jmoiron/sqlx/blob/41dac167fdad5e3fd81d66cafba0951dc6823a30/bind.go#L60 (simplified version)
func rebind(query string) string {
	rqb := make([]byte, 0)

	var i, j int

	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		rqb = append(rqb, query[:i]...)
		rqb = append(rqb, '$')
		j++
		rqb = strconv.AppendInt(rqb, int64(j), 10)
		query = query[i+1:]
	}

	return string(append(rqb, query...))
}
