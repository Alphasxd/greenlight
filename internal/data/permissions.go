package data

import (
	"context"
	"database/sql"
	"time"
)

// Permissions 定义了用户可以拥有的权限列表
type Permissions []string

// Include 检查给定的权限代码是否在 Permissions 中
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *sql.DB
}

// GetAllForUser 返回指定用户的权限列表
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
        SELECT permissions.code
        FROM permissions
        INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
        INNER JOIN users ON users_permissions.user_id = users.id
        WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	var permissions Permissions

	// 循环遍历结果集中的行，rows.Next() 方法返回 bool 值，指示是否还有更多的行可以迭代
	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	// 检查 rows.Next() 循环过程中是否有错误
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
