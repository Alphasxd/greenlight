package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordNotFound 表示 Get() 方法未找到记录的错误。
	ErrRecordNotFound = errors.New("record not found")
	// ErrEditConflict 表示存在编辑冲突的错误。
	ErrEditConflict = errors.New("edit conflict")
)

// Models 定义一个模型结构体，包含所有模型的实例
type Models struct {
	Movies      MovieModel
	Tokens      TokenModel
	Users       UserModel
	Permissions PermissionModel
}

// NewModels 函数返回一个包含所有模型的 Models 结构体实例
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}
