// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
)

type ResultSelector struct {
	Filter  *filter.Request  `json:"filter" binding:"required"`
	Sorting *sorting.Request `json:"sorting" binding:"omitempty"`
	Paging  *paging.Request  `json:"paging" binding:"omitempty"`
}
