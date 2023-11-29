package paging

type PagingSettingsInterface interface {
	GetPagingDefault() (pageSize int)
}
