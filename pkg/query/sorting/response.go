package sorting

type Response struct {
	SortingColumn    string        `json:"column"`
	SortingDirection SortDirection `json:"direction"`
}
