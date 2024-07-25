// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package paging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestPagingModel struct {
	ID   int    `sortColumn:"TestModel.ID" shownName:"ID" sortDirection:"desc"`
	Name string `sortColumn:"name" shownName:"TheName"`
}

func (t TestPagingModel) GetPagingDefault() (pageSize int) {
	return 20
}

type FailedTestPagingModel struct {
	ID   int
	Name string
}

func (f FailedTestPagingModel) GetPagingDefault() (pageSize int) {
	return 0
}

func TestPagingRules(t *testing.T) {
	testPagingModelRequest := &Request{
		PageIndex: 0,
		PageSize:  20,
	}

	inValidPagingReq := &Request{PageIndex: 0, PageSize: 0}

	t.Run("negative paging request", func(t *testing.T) {
		invalidReq := &Request{PageIndex: -1, PageSize: 0}
		request, err := ValidateAndApplyPagingRules(&TestPagingModel{}, invalidReq)
		assert.NoError(t, err)
		assert.Equal(t, testPagingModelRequest, request)
	})

	t.Run("invalid paging Request", func(t *testing.T) {
		request, vErr := ValidateAndApplyPagingRules(&TestPagingModel{}, inValidPagingReq)
		assert.NoError(t, vErr)
		assert.EqualValues(t, request, testPagingModelRequest)
	})

	t.Run("not tagged struct paging Request", func(t *testing.T) {
		request, vErr := ValidateAndApplyPagingRules(&FailedTestPagingModel{}, inValidPagingReq)
		assert.NoError(t, vErr)
		assert.EqualValues(t, request, testPagingModelRequest)
	})
}
