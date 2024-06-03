// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import (
	"strconv"
	"time"

	esextensions "github.com/greenbone/opensight-golang-libraries/pkg/openSearch/esextension"

	"github.com/aquasecurity/esquery"
	"github.com/samber/lo"
)

// HandleCompareOperatorIsEqualTo handles is equal to
func HandleCompareOperatorIsEqualTo(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return createTermQuery(fieldName, fieldValue, fieldKeys, querySettings)
}

// HandleCompareOperatorIsKeywordEqualTo handles is keyword field equal to
func HandleCompareOperatorIsKeywordEqualTo(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return createTermQuery(fieldName+".keyword", fieldValue, fieldKeys, querySettings)
}

// HandleCompareOperatorContains handles contains
func HandleCompareOperatorContains(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	if querySettings.UseNestedMatchQueryFields != nil &&
		querySettings.UseNestedMatchQueryFields[fieldName] {
		return nestedHandleCompareOperatorContains(fieldName, fieldKeys, fieldValue, querySettings)
	}
	// for list of values
	if querySettings.WildcardArrays != nil &&
		querySettings.WildcardArrays[fieldName] {
		return handleCompareOperatorContainsDifferent(fieldName, nil, fieldValue, querySettings)
	} else {
		if values, ok := fieldValue.([]interface{}); ok {
			return esquery.Bool().
				Should(
					lo.Map[interface{}, esquery.Mappable](values, func(value interface{}, _ int) esquery.Mappable {
						return esquery.Wildcard(fieldName+".keyword", "*"+valueToString(value)+"*")
					})...,
				).
				MinimumShouldMatch(1)
		} else { // for single values
			return esquery.Wildcard(fieldName+".keyword", "*"+valueToString(fieldValue)+"*")
		}
	}
}

func nestedHandleCompareOperatorContains(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	// Special case as now for one input we need to queries to be set (for name and value)
	nestedFieldSetting := findNestedFieldByName(fieldName, querySettings)
	if nestedFieldSetting != nil && len(fieldKeys) == 1 {
		query1 := esextensions.Nested(nestedFieldSetting.FieldKeyName, *esquery.Bool().
			Must(
				esquery.Match(nestedFieldSetting.FieldKeyName, fieldKeys[0]),
				esquery.Wildcard(nestedFieldSetting.FieldValueName, "*"+valueToString(fieldValue)+"*")))
		return query1
	}
	return nil
}

func handleCompareOperatorContainsDifferent(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	if values, ok := fieldValue.([]interface{}); ok {
		return esquery.Bool().
			Should(
				lo.Map[interface{}, esquery.Mappable](values, func(value interface{}, _ int) esquery.Mappable {
					return esquery.Wildcard(fieldName, "*"+valueToString(value)+"*")
				})...,
			).
			MinimumShouldMatch(1)
	} else { // for single values
		return esquery.Wildcard(fieldName, "*"+fieldValue.(string)+"*")
	}
}

// HandleCompareOperatorBeginsWith handles begins with
func HandleCompareOperatorBeginsWith(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	isWildcardArray := querySettings.WildcardArrays != nil && querySettings.WildcardArrays[fieldName]
	field := fieldName + ".keyword"

	// if the query settings specify that it is a wildcard array, use the field name appended with the '.keyword'
	if isWildcardArray {
		field = fieldName
	}

	return handleCompareOperatorBeginsWith(field, fieldValue)
}

func handleCompareOperatorBeginsWith(fieldName string, fieldValue any) esquery.Mappable {
	if values, ok := fieldValue.([]any); ok {
		// for a list of values
		return esquery.Bool().
			Should(
				lo.Map[any, esquery.Mappable](values, func(value any, _ int) esquery.Mappable {
					return esquery.Prefix(fieldName, valueToString(value))
				})...,
			).
			MinimumShouldMatch(1)
	}

	// for single value
	return esquery.Prefix(fieldName, valueToString(fieldValue))
}

// HandleCompareOperatorNotBeginsWith handles not begins with
func HandleCompareOperatorNotBeginsWith(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	// for list of values
	if values, ok := fieldValue.([]interface{}); ok {
		return esquery.Bool().
			MustNot(
				lo.Map[interface{}, esquery.Mappable](values, func(value interface{}, _ int) esquery.Mappable {
					return esquery.Prefix(fieldName+".keyword", valueToString(value))
				})...,
			)
	} else { // for single values
		return esquery.Prefix(fieldName+".keyword", fieldValue.(string))
	}
}

// HandleCompareOperatorIsLessThanOrEqualTo handles is less than or equal to
func HandleCompareOperatorIsLessThanOrEqualTo(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return esquery.Range(fieldName).
		Lte(fieldValue)
}

// HandleCompareOperatorIsGreaterThanOrEqualTo handles is greater than or equal to
func HandleCompareOperatorIsGreaterThanOrEqualTo(fieldName string, fieldKeys []string,
	fieldValue any, querySettings *QuerySettings,
) esquery.Mappable {
	return esquery.Range(fieldName).
		Gte(fieldValue)
}

// HandleCompareOperatorIsGreaterThan handles is greater than
func HandleCompareOperatorIsGreaterThan(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return esquery.Range(fieldName).
		Gt(fieldValue)
}

// HandleCompareOperatorIsLessThan handles is less than
func HandleCompareOperatorIsLessThan(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	return esquery.Range(fieldName).
		Lt(fieldValue)
}

func simpleNestedMatchQuery(fieldName string, fieldKeys []string, fieldValue any, querySettings *QuerySettings) esquery.Mappable {
	// Special case as now for one input we need to queries to be set (for name and value)
	nestedFieldSetting := findNestedFieldByName(fieldName, querySettings)
	if nestedFieldSetting != nil && len(fieldKeys) == 1 {
		query1 := esextensions.Nested(nestedFieldSetting.FieldName, *esquery.Bool().Must(
			esquery.Match(nestedFieldSetting.FieldKeyName, fieldKeys[0]),
			esquery.Match(nestedFieldSetting.FieldValueName, fieldValue)))
		return query1
	}
	return nil
}

func createTermQuery(fieldName string, fieldValue any, fieldKeys []string, querySettings *QuerySettings) esquery.Mappable {
	if querySettings.UseNestedMatchQueryFields != nil &&
		querySettings.UseNestedMatchQueryFields[fieldName] {
		return simpleNestedMatchQuery(fieldName, fieldKeys, fieldValue, querySettings)
	}
	// for list of values
	if values, ok := fieldValue.([]interface{}); ok {
		if len(values) == 0 {
			return nil
		}

		if querySettings.IsEqualToKeywordFields != nil &&
			querySettings.IsEqualToKeywordFields[fieldName] {
			fieldName = fieldName + ".keyword"
		}
		if querySettings.UseMatchPhrase != nil &&
			querySettings.UseMatchPhrase[fieldName] {
			return esquery.MatchPhrase(fieldName, values...)
		}
		return esquery.Terms(fieldName, values...)
	} else { // for single values
		if querySettings.UseMatchPhrase != nil &&
			querySettings.UseMatchPhrase[fieldName] {
			return esquery.MatchPhrase(fieldName, fieldValue)
		}
		return esquery.Term(fieldName, fieldValue)
	}
}

func valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return ""
	}
}

func findNestedFieldByName(name string, querySettings *QuerySettings) *NestedQueryFieldDefinition {
	if querySettings.NestedQueryFieldDefinitions == nil {
		return nil
	}
	for _, field := range querySettings.NestedQueryFieldDefinitions {
		if field.FieldName == name {
			return &field // Gibt die Adresse des gefundenen Feldes zurück
		}
	}
	return nil // Kein passendes Feld gefunden
}
