// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package sorting

type SortingSettingsInterface interface {
	GetSortDefault() SortDefault
	GetSortingMap() map[string]string
	GetOverrideSortColumn(string) string
}
