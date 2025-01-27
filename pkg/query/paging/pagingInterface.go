// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package paging

type PagingSettingsInterface interface {
	GetPagingDefault() (pageSize int)
}
