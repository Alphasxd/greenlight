package data

import "github.com/Alphasxd/greenlight/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {

	// 检查 Page 字段
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")

	// 检查 PageSize 字段
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// 检查 Sort 字段
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
