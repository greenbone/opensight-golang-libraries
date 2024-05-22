// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package query facilitates the translation of a result selector into a PostgresSQL conditional query string, incorporating sorting and paging functionalities.
package query

import (
	"fmt"
	"strings"

	"github.com/greenbone/opensight-golang-libraries/pkg/query"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
	"github.com/pkg/errors"
)

// Settings is a configuration struct used to customize the behavior of the query builder.
type Settings struct {
	FilterFieldMapping map[string]string // Mapping of filter fields for query customization
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

// AddFilterRequest appends filter conditions to the query builder based on the provided filter request.
// It constructs conditional clauses using the logic operator specified in the request.
// TODO: Enhance the AddFilterRequest function to prevent SQL injection vulnerabilities.
// AddFilterRequest is currently vulnerable to sql injection and should be used only when the input is trusted
func (qb *Builder) AddFilterRequest(request *filter.Request) error {
	if request == nil || len(request.Fields) == 0 {
		return nil
	}
	logicOperator := strings.ToUpper(string(request.Operator))

	qb.query.WriteString("WHERE")

	for index, field := range request.Fields {
		var (
			err               error
			valueIsList       bool
			conditionTemplate string
			conditionParams   []any
		)
		valueIsList, _, err = checkFieldValueType(field)
		if err != nil {
			return fmt.Errorf("error checking filter field value type '%s': %w", field, err)
		}
		conditionTemplate, conditionParams, err =
			composeQuery(qb.querySettings.FilterFieldMapping, field, valueIsList)
		if err != nil {
			return fmt.Errorf("error composing query from filter field %w", err)
		}
		if index > 0 {
			qb.query.WriteString(fmt.Sprintf(" %s", logicOperator))
		}
		qb.query.WriteString(generateConditionalQuery(conditionParams, conditionTemplate))
	}
	return nil
}

// generateConditionalQuery generates filter conditions based on the provided filter field value,
// field mappings, logic operators, and a flag indicating if these are the final request fields.
// If the field value is not of the expected type or cannot be converted to a string, it skips that field.
// The finalRequestFields flag determines whether to append the logic operator at the end of the query.
func generateConditionalQuery(conditionParams []any, column string) string {
	var (
		queryString strings.Builder
		conditions  []string
	)

	queryString.WriteString(fmt.Sprintf(" %s (", column))
	for _, conditionParam := range conditionParams {
		params, ok := conditionParam.([]any)
		if !ok {
			continue
		}
		for _, condition := range params {
			conditions = append(conditions, fmt.Sprintf("'%s'", condition))
		}
	}
	conditionalQuery := strings.Join(conditions, ", ")
	queryString.WriteString(fmt.Sprintf("%s)", conditionalQuery))
	return queryString.String()
}

// AddSorting appends sorting conditions to the query builder based on the provided sorting request.
// It constructs the ORDER BY clause using the specified sort column and direction.
func (qb *Builder) AddSorting(sort *sorting.Request) error {
	if sort == nil {
		return errors.New("missing sorting fields, add sort request or remove call to AddSort()")
	}

	// map fields to column
	sortColumn, ok := qb.querySettings.FilterFieldMapping[sort.SortColumn]
	if !ok {
		return fmt.Errorf("mapping for sort column '%s' has not been implemented", sort.SortColumn)
	}

	qb.query.WriteString(fmt.Sprintf(" ORDER BY %s %s", sortColumn, sort.SortDirection))
	return nil
}

// AddPaging appends paging conditions to the query builder based on the provided paging request.
// It constructs the OFFSET and LIMIT clauses according to the specified page index and page size.
func (qb *Builder) AddPaging(paging *paging.Request) error {
	if paging == nil {
		return errors.New("missing paging fields, add paging request or remove call to AddSize()")
	}

	if paging.PageIndex > 1 {
		qb.query.WriteString(fmt.Sprintf(" OFFSET %d", paging.PageIndex))
	}

	if paging.PageSize > 0 {
		qb.query.WriteString(fmt.Sprintf(" LIMIT %d", paging.PageSize))
	}
	return nil
}

// Build generates the complete SQL query based on the provided result selector.
// It constructs the query by adding filter, sorting, and paging conditions.
// If any error occurs during the construction, it returns an empty string.
func (qb *Builder) Build(resultSelector query.ResultSelector) string {
	if resultSelector.Filter != nil {
		err := qb.AddFilterRequest(resultSelector.Filter)
		if err != nil {
			return ""
		}
	}

	if resultSelector.Sorting != nil {
		err := qb.AddSorting(resultSelector.Sorting)
		if err != nil {
			return ""
		}
	}

	if resultSelector.Paging != nil {
		err := qb.AddPaging(resultSelector.Paging)
		if err != nil {
			return ""
		}
	}

	return qb.query.String()
}
