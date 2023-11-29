package sorting

type Request struct {
	SortColumn    string        `json:"column"`
	SortDirection SortDirection `json:"direction"`
}
