package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordNotFound 表示 Get() 方法未找到记录的错误。
	ErrRecordNotFound = errors.New("record not found")
)

// Models 定义一个模型结构体，包含所有模型的实例
type Models struct {
	Movies MovieModel
}

// NewModels 函数返回一个包含所有模型的 Models 结构体实例
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
