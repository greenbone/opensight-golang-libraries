package paging

import (
	"gorm.io/gorm"
)

type Request struct {
	PageIndex int `json:"index"`
	PageSize  int `json:"size"`
}

// AddRequest adds pagination to the gorm transaction based on the given request.
//
// transaction: The GORM database transaction.
// request: The request object containing pagination information.
//   - PageIndex: The index of the page to retrieve (starting from 0).
//   - PageSize: The number of records to retrieve per page.
//
// Returns the modified transaction with the pagination applied.
func AddRequest(transaction *gorm.DB, request *Request) *gorm.DB {
	offset := 0
	if request.PageIndex >= 0 {
		offset = request.PageIndex * request.PageSize
		transaction = transaction.Offset(offset)
	}
	transaction = transaction.Limit(request.PageSize)
	return transaction
}
