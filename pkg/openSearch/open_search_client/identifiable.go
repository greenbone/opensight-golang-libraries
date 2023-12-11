// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_client

type Identifiable interface {
	GetId() string

	SetId(id string)
}
