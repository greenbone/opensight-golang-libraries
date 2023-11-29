package sorting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestSortingModel struct {
	ID   int    `displayName:"ID"`
	Name string `displayName:"TheName" pagesize:"20"`
}

func (t TestSortingModel) GetSortDefault() SortDefault {
	return SortDefault{
		Column:    "TestModel.ID",
		Direction: DirectionDescending,
	}
}

func (t TestSortingModel) GetSortingMap() map[string]string {
	return map[string]string{
		"name":  "name",
		"test2": "test2",
	}
}

func (t TestSortingModel) GetOverrideSortColumn(s string) string {
	return "\"appliance\".\"name\""
}

func (t TestSortingModel) GetFilterMapping(inputString string) string {
	return "test"
}

func (t TestSortingModel) GetFilterMap() map[string]string {
	return map[string]string{
		"test":  "test",
		"test2": "test2",
	}
}

func TestSortingRule(t *testing.T) {
	defaultPagingRequest := &Request{
		SortColumn:    "TestModel.ID",
		SortDirection: "desc",
	}
	invalidSortingField := &Request{SortColumn: "InvalidField", SortDirection: DirectionAscending}

	t.Run("invalid sorting field", func(t *testing.T) {
		request, vErr := DetermineEffectiveSortingParams(&TestSortingModel{}, invalidSortingField)
		assert.NoError(t, vErr)
		assert.EqualValues(t, Params{
			OriginalSortColumn:  defaultPagingRequest.SortColumn,
			SortDirection:       defaultPagingRequest.SortDirection,
			EffectiveSortColumn: defaultPagingRequest.SortColumn,
		}, request)
	})

	t.Run("valid sorting field", func(t *testing.T) {
		validRequest := &Request{SortColumn: "TestModel.ID", SortDirection: DirectionDescending}
		finalReq, vErr := DetermineEffectiveSortingParams(&TestSortingModel{}, validRequest)
		assert.NoError(t, vErr)
		assert.EqualValues(t, Params{
			OriginalSortColumn:  validRequest.SortColumn,
			SortDirection:       validRequest.SortDirection,
			EffectiveSortColumn: validRequest.SortColumn,
		}, finalReq)
	})

	t.Run("valid sorting field with override", func(t *testing.T) {
		validRequest := &Request{SortColumn: "name", SortDirection: DirectionDescending}
		finalReq, vErr := DetermineEffectiveSortingParams(&TestSortingModel{}, validRequest)
		assert.NoError(t, vErr)
		assert.EqualValues(t, Params{
			OriginalSortColumn:  validRequest.SortColumn,
			SortDirection:       validRequest.SortDirection,
			EffectiveSortColumn: "\"appliance\".\"name\"",
		}, finalReq)
	})

	t.Run("sorting direction asc", func(t *testing.T) {
		assert.EqualValues(t, DirectionDescending.String(), "DESC")
	})

	t.Run("sorting direction desc", func(t *testing.T) {
		assert.EqualValues(t, DirectionAscending.String(), "ASC")
	})

	t.Run("sorting direction none", func(t *testing.T) {
		assert.EqualValues(t, NoDirection.String(), "")
	})

	t.Run("get sorting direction string asc", func(t *testing.T) {
		assert.EqualValues(t, SortDirectionFromString("asc"), DirectionAscending)
	})

	t.Run("get sorting direction string desc", func(t *testing.T) {
		assert.EqualValues(t, SortDirectionFromString("desc"), DirectionDescending)
	})

	t.Run("get no sorting direction", func(t *testing.T) {
		assert.EqualValues(t, SortDirectionFromString(""), NoDirection)
	})
}
