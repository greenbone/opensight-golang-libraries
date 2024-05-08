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
func (qb *Builder) AddFilterRequest(request *filter.Request) error {
	if request == nil {
		return errors.New("missing filter request, add filter request body or remove the call to AddFilterRequest()")
	}

	logicOperator := strings.ToUpper(string(request.Operator))

	qb.query.WriteString("WHERE ")

	filterFieldsLen := len(request.Fields)
	for index, field := range request.Fields {
		operator, err := mapCompareOperatorToSQLOperator(string(field.Operator))
		if err != nil {
			return err
		}
		mappedField, err := mapFilterFieldToEntityName(field, qb.querySettings.FilterFieldMapping)
		if err != nil {
			return err
		}
		qb.query.WriteString(generateConditionalQuery(field.Value, mappedField, operator,
			logicOperator, index == filterFieldsLen-1))
	}
	return nil
}

// generateConditionalQuery generates filter conditions based on the provided filter field value,
// field mappings, logic operators, and a flag indicating if these are the final request fields.
// If the field value is not of the expected type or cannot be converted to a string, it skips that field.
// The finalRequestFields flag determines whether to append the logic operator at the end of the query.
func generateConditionalQuery(filterFieldValue any, fieldMappings, mappedLogicOperator, logicOperator string, finalRequestFields bool) string {
	var queryString strings.Builder

	fieldList, ok := filterFieldValue.([]any)
	if !ok {
		return ""
	}
	fieldLen := len(fieldList)
	for index, field := range fieldList {
		fieldString, ok := field.(string)
		if !ok {
			continue
		}
		if index != fieldLen-1 || !finalRequestFields {
			queryString.WriteString(fmt.Sprintf("%s %s '%s' ", fieldMappings, mappedLogicOperator, fieldString))
		} else {
			queryString.WriteString(fmt.Sprintf("%s %s %s '%s' ", logicOperator,
				fieldMappings, mappedLogicOperator, fieldString))
		}
	}
	return queryString.String()
}

// AddSorting appends sorting conditions to the query builder based on the provided sorting request.
// It constructs the ORDER BY clause using the specified sort column and direction.
func (qb *Builder) AddSorting(sort *sorting.Request) error {
	if sort == nil {
		return errors.New("missing sorting fields, add sort request or remove call to AddSort()")
	}
	qb.query.WriteString(fmt.Sprintf("ORDER BY %s %s ", sort.SortColumn, sort.SortDirection))
	return nil
}

// AddPaging appends paging conditions to the query builder based on the provided paging request.
// It constructs the OFFSET and LIMIT clauses according to the specified page index and page size.
func (qb *Builder) AddPaging(paging *paging.Request) error {
	if paging == nil {
		return errors.New("missing paging fields, add paging request or remove call to AddSize()")
	}

	if paging.PageIndex > 1 {
		qb.query.WriteString(fmt.Sprintf("OFFSET %d ", paging.PageIndex))
	}

	if paging.PageSize > 0 {
		qb.query.WriteString(fmt.Sprintf("LIMIT %d", paging.PageSize))
	}
	return nil
}

// mapFilterFieldToEntityName retrieves the mapped entity name for the given DTO field from the provided field mapping.
// If the mapping is not found, it returns an error indicating that the mapping for the filter field is not implemented.
func mapFilterFieldToEntityName(dtoField filter.RequestField, fieldMapping map[string]string) (string, error) {
	entityName, ok := fieldMapping[dtoField.Name]
	if !ok {
		return "", filter.NewInvalidFilterFieldError(
			"Mapping for filter field '%s' is currently not implemented.", dtoField.Name)
	}
	return entityName, nil
}

// mapCompareOperatorToSQLOperator maps the given compare operator string to its corresponding SQL operator.
// It looks up the mapping using the compareOperator as the key in the mapping map.
// If the mapping is not found, it returns an error indicating that the compare operator is unknown.
func mapCompareOperatorToSQLOperator(compareOperator string) (string, error) {
	mapping := map[string]string{
		string(filter.CompareOperatorIsEqualTo):    "=",
		string(filter.CompareOperatorIsNotEqualTo): "!=",
	}

	operator, ok := mapping[compareOperator]
	if !ok {
		return "", fmt.Errorf("unknown compare openrator %s", compareOperator)
	}
	return operator, nil
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
