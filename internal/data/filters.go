package data

import (
	"math"
	"strings"

	"github.com/Alphasxd/greenlight/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// ValidateFilters 方法检查过滤器字段是否包含有效的值。如果有错误，方法会将错误添加到 v.Errors 中。
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

// sortColumn 方法返回排序列的名称，即数据库中的列名。
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	// 直接 panic，防止 SQL 注入攻击
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection 方法返回排序方向，即 ASC 或 DESC。
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC" // descending
	}
	return "ASC" // ascending
}

// limit 方法返回 LIMIT 子句的值。
func (f Filters) limit() int {
	return f.PageSize
}

// offset 方法返回 OFFSET 子句的值。
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// calculateMetadata 方法返回一个包含了元数据的 Metadata 结构体。
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))), // ceil 向上取整 ceiling 天花板
		TotalRecords: totalRecords,
	}
}
