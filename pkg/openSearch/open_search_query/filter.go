// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_query

import "github.com/greenbone/opensight-golang-libraries/pkg/query/filter"

var filterFieldMapping = map[string]string{
	"name":            "vulnerabilityTest.name",
	"operatingSystem": "asset.operatingSystem",
	"hostname":        "asset.hostnames",
	"ip":              "asset.ips",
	"macAddresses":    "asset.macAddresses",
	"severity":        "vulnerabilityTest.severityCvss.override",
	"qod":             "qod",
	"host":            "host.name",
	"port":            "port",
	"solutionType":    "vulnerabilityTest.solutionType",
	"cve":             "vulnerabilityTest.references.cve",
	"oid":             "vulnerabilityTest.oid",
	"family":          "vulnerabilityTest.family",
	"lastScanAt":      "asset.last_scan_date",
	"id":              "asset.id",
	"applianceName":   "appliance.name",
	"applianceId":     "appliance.id",
	"tag":             "asset.tags",
	"tagName":         "asset.tags.tagname",
	"tagValue":        "asset.tags.tagvalue",
}

func EffectiveFilterFields(filterRequest filter.Request) (filter.Request, error) {
	var filterFields []filter.RequestField
	for _, field := range filterRequest.Fields {
		mappedField, err := createMappedField(field)
		if err != nil {
			return filter.Request{}, err
		}
		filterFields = append(filterFields, mappedField)
	}
	return filter.Request{
		Operator: filterRequest.Operator,
		Fields:   filterFields,
	}, nil
}

func createMappedField(dtoField filter.RequestField) (filter.RequestField, error) {
	entityName, ok := filterFieldMapping[dtoField.Name]
	if !ok {
		return filter.RequestField{}, filter.NewInvalidFilterFieldError(
			"Mapping for filter field '%s' is currently not implemented.", dtoField.Name)
	}

	return filter.RequestField{
		Operator: dtoField.Operator,
		Keys:     dtoField.Keys,
		Name:     entityName,
		Value:    dtoField.Value,
	}, nil
}
