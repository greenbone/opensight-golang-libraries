package paging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagingValidations(t *testing.T) {
	t.Run("shallRaiseNegativError", func(t *testing.T) {
		req := &Request{
			PageIndex: -1,
			PageSize:  0,
		}

		err := validatePagingRequest(req)
		assert.ErrorContains(t, err, "page index must be 0 or greater than 0")
	})

	t.Run("shallRaiseNilError", func(t *testing.T) {
		err := validatePagingRequest(nil)
		assert.ErrorContains(t, err, "request is nil")
	})

	t.Run("shallRaisePage0Error", func(t *testing.T) {
		req := &Request{
			PageIndex: 1,
			PageSize:  0,
		}

		err := validatePagingRequest(req)
		assert.ErrorContains(t, err, "page size must be greater than 0")
	})

	t.Run("shallHaveNotPagingError", func(t *testing.T) {
		req := &Request{
			PageIndex: 1,
			PageSize:  10,
		}

		err := validatePagingRequest(req)
		assert.NoError(t, err)
	})
}
