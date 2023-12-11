package sorting

// Response represents the response structure for sorting column and direction.
// SortingColumn stores the name of the column which was used for sorting.
// SortingDirection stores the direction which was applied by the sorting.
type Response struct {
	SortingColumn    string        `json:"column"`
	SortingDirection SortDirection `json:"direction"`
}
