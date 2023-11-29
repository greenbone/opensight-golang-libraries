package sorting

type SortingSettingsInterface interface {
	GetSortDefault() SortDefault
	GetSortingMap() map[string]string
	GetOverrideSortColumn(string) string
}
