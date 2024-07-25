// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package sorting

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	// SortColumnTag is a tag to define the field name to be used for sorting
	SortColumnTag = "sortColumn"
	/*SortColumnOverrideTag is a tag to override the field name to be used for sorting.

	An example use case is the following:

	If you take a column to order by, which is part of the original table (like hostname) the SQL syntax is fine - see the generated GORM query below.´.
	SELECT "asset"."id","asset"."created_at","asset"."updated_at","asset"."greenbone_agent_id","asset"."last_authenticated_scan_at","asset"."mac_address","asset"."net_bios_name","asset"."ssh_fingerprint","asset"."has_agent","asset"."has_vt_result","asset"."hostname","asset"."ip","asset"."last_scan_at","asset"."operating_system","asset"."deleted_at","asset"."deleted_by","asset"."source_id","asset"."appliance_id" FROM "asset" LEFT JOIN "appliance" "Appliance" ON "asset"."appliance_id" = "Appliance"."id" LEFT JOIN "installed_software" "InstalledSoftwares" ON "asset"."id" = "InstalledSoftwares"."asset_id" WHERE ("Appliance"."name" ILIKE '%Example%') ORDER BY hostname ASC LIMIT 10

	If we now change to sort by a column of a joined table, the SQL syntax is "TABLENAME"."Columnname". Now see the
	LEFT JOIN "appliance" "Appliance" which means to access the table appliance we now need to use "Appliance" instead of "appliance".
	Therefore its necassary to change the SQL syntax for the foreign table sort.
	SELECT "asset"."id","asset"."created_at","asset"."updated_at","asset"."greenbone_agent_id","asset"."last_authenticated_scan_at","asset"."mac_address","asset"."net_bios_name","asset"."ssh_fingerprint","asset"."has_agent","asset"."has_vt_result","asset"."hostname","asset"."ip","asset"."last_scan_at","asset"."operating_system","asset"."deleted_at","asset"."deleted_by","asset"."source_id","asset"."appliance_id" FROM "asset" LEFT JOIN "appliance" "Appliance" ON "asset"."appliance_id" = "Appliance"."id" LEFT JOIN "installed_software" "InstalledSoftwares" ON "asset"."id" = "InstalledSoftwares"."asset_id" ORDER BY "Appliance"."name" DESC LIMIT 20

	If we query for the appliance name GORM changes it correctly to "Appliance"."name" - see the query below. And the same we now do in our code for the sorting - we change the JOINED field from appliance.name to "Appliance"."name".
	SELECT "asset"."id","asset"."created_at","asset"."updated_at","asset"."greenbone_agent_id","asset"."last_authenticated_scan_at","asset"."mac_address","asset"."net_bios_name","asset"."ssh_fingerprint","asset"."has_agent","asset"."has_vt_result","asset"."hostname","asset"."ip","asset"."last_scan_at","asset"."operating_system","asset"."deleted_at","asset"."deleted_by","asset"."source_id","asset"."appliance_id" FROM "asset" LEFT JOIN "appliance" "Appliance" ON "asset"."appliance_id" = "Appliance"."id" LEFT JOIN "installed_software" "InstalledSoftwares" ON "asset"."id" = "InstalledSoftwares"."asset_id" WHERE ("Appliance"."name" ILIKE '%Example%') ORDER BY hostname ASC LIMIT 10
	*/
	SortColumnOverrideTag = "sortColumnOverride"
	// SortDirectionTag is a tag that defines the sort (i.e. direction of sorting); must be a SortDirection
	SortDirectionTag = "sortDirection"
)

func AddRequest(transaction *gorm.DB, params Params) *gorm.DB {
	if params.SortDirection != NoDirection && params.EffectiveSortColumn != "" {
		transaction = transaction.Order(params.EffectiveSortColumn + " " + params.SortDirection.String())
	}
	return transaction
}

// SortableColumn is a struct to hold the fields which the paging can sort by.
type SortableColumn struct {
	Column         string
	ColumnOverride string
}

// SortDefault holds the default for sort direction and sorting field.
type SortDefault struct {
	Column    string
	Direction SortDirection
}

// GetSortDefaults returns the sortable fields based on the struct provided.
func GetSortDefaults(model SortingSettingsInterface) (result SortDefault, err error) {
	return model.GetSortDefault(), nil
}

// GetSortableColumns returns a list of sortable fields
func GetSortableColumns(model SortingSettingsInterface) (sortables []SortableColumn) {
	sortingMap := model.GetSortingMap()
	for column := range sortingMap {
		override := model.GetOverrideSortColumn(sortingMap[column])
		sortables = append(sortables, SortableColumn{
			Column:         column,
			ColumnOverride: override,
		})
	}
	return sortables
}

func fieldIsSortable(model SortingSettingsInterface, field string) bool {
	sortables := GetSortableColumns(model)
	for _, sortableField := range sortables {
		if sortableField.Column == field {
			return true
		}
	}
	return false
}

func getEffectiveSortColumn(model SortingSettingsInterface, field string) string {
	sortables := GetSortableColumns(model)
	for _, sortableField := range sortables {
		if sortableField.Column == field {
			if sortableField.ColumnOverride != "" {
				return sortableField.ColumnOverride
			} else {
				return sortableField.Column
			}
		}
	}
	return field
}

// DetermineEffectiveSortingParams checks the requested sorting and sets the defaults in case of an error.
// If a SortColumnOverrideTag (sortColumnOverride) is given, it's value will be used for sorting instead
// of SortColumnTag (sortColumn). For a detailed explanation see SortColumnOverrideTag
func DetermineEffectiveSortingParams(model SortingSettingsInterface, sortingReq *Request) (Params, error) {
	// Validate the sorting request
	if err := ValidateSortingRequest(sortingReq); err == nil {
		if fieldIsSortable(model, sortingReq.SortColumn) {
			params := paramsOf(*sortingReq)
			params.EffectiveSortColumn = getEffectiveSortColumn(model, sortingReq.SortColumn)
			return params, nil
		}
	}

	// If there is an error in the request for sorting, get the defaults and apply them.
	sortingDefaults, pdErr := GetSortDefaults(model)
	if pdErr != nil {
		err := fmt.Errorf("failed to get sorting defaults: %w", pdErr)
		sortingReq.SortColumn = ""
		sortingReq.SortDirection = DirectionDescending
		return paramsOf(*sortingReq), err
	}

	if sortingReq == nil {
		sortingReq = &Request{}
	}
	sortingReq.SortColumn = sortingDefaults.Column
	sortingReq.SortDirection = sortingDefaults.Direction
	return paramsOf(*sortingReq), nil
}

func paramsOf(sortingReq Request) Params {
	return Params{
		OriginalSortColumn: sortingReq.SortColumn,
		SortDirection:      sortingReq.SortDirection, EffectiveSortColumn: sortingReq.SortColumn,
	}
}
