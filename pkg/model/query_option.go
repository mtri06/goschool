package model

type Operator string

type Pagination struct {
	Page     int
	PageSize int
}

func NewPagination(page, pageSize int) Pagination {
	return Pagination{Page: page, PageSize: pageSize}
}

type Order struct {
	Field string
	Desc  bool
}

type OrderBy []Order
