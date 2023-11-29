package sorting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortingValidations(t *testing.T) {
	t.Run("shallRaiseNoSortingError", func(t *testing.T) {
		req := &Request{
			SortColumn:    "col1",
			SortDirection: DirectionAscending,
		}

		err := ValidateSortingRequest(req)
		assert.NoError(t, err)
	})

	t.Run("shallRaiseSortingColumnError", func(t *testing.T) {
		req := &Request{
			SortColumn:    "",
			SortDirection: DirectionAscending,
		}

		err := ValidateSortingRequest(req)
		assert.ErrorContains(t, err, "sorting column is empty")
	})

	t.Run("shallRaiseSortingDirectionError", func(t *testing.T) {
		req := &Request{
			SortColumn:    "col1",
			SortDirection: "",
		}

		err := ValidateSortingRequest(req)
		assert.ErrorContains(t, err, "sorting direction is empty")
	})

	t.Run("shallRaiseSortingDirectionError-invalidSortDirection", func(t *testing.T) {
		req := &Request{
			SortColumn:    "col1",
			SortDirection: "nonsense",
		}

		err := ValidateSortingRequest(req)
		assert.ErrorContains(t, err, "NONSENSE is no valid sorting direction, possible values are asc, desc")
	})
}
