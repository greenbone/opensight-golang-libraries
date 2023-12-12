// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchQuery

import queryFilter "github.com/greenbone/opensight-golang-libraries/pkg/query/filter"

func EffectiveFilterFields(filterRequest queryFilter.Request, fieldMapping map[string]string) (queryFilter.Request, error) {
	var filterFields []queryFilter.RequestField
	for _, field := range filterRequest.Fields {
		mappedField, err := createMappedField(field, fieldMapping)
		if err != nil {
			return queryFilter.Request{}, err
		}
		filterFields = append(filterFields, mappedField)
	}
	return queryFilter.Request{
		Operator: filterRequest.Operator,
		Fields:   filterFields,
	}, nil
}

func createMappedField(dtoField queryFilter.RequestField, fieldMapping map[string]string) (queryFilter.RequestField, error) {
	entityName, ok := fieldMapping[dtoField.Name]
	if !ok {
		return queryFilter.RequestField{}, queryFilter.NewInvalidFilterFieldError(
			"Mapping for filter field '%s' is currently not implemented.", dtoField.Name)
	}

	return queryFilter.RequestField{
		Operator: dtoField.Operator,
		Keys:     dtoField.Keys,
		Name:     entityName,
		Value:    dtoField.Value,
	}, nil
}
