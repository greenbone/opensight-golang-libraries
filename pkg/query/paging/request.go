package paging

import (
	"gorm.io/gorm"
)

type Request struct {
	PageIndex int `json:"index"`
	PageSize  int `json:"size"`
}

// AddRequest adds the paging and sorting information to the transaction
func AddRequest(transaction *gorm.DB, request *Request) *gorm.DB {
	offset := 0
	if request.PageIndex >= 0 {
		offset = request.PageIndex * request.PageSize
		transaction = transaction.Offset(offset)
	}
	transaction = transaction.Limit(request.PageSize)
	return transaction
}
