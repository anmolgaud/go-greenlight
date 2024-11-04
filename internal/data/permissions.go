package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Permissions []string

func (p Permissions) Include(code string) bool {
	for _, v := range(p) {
		if code == v {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *sql.DB
}

func (m *PermissionModel) GetAllForUser(userId int64) (Permissions, error) {
	query := `
	SELECT pm.code
	FROM permissions pm
	INNER JOIN user_permissions upm
	ON pm.id = upm.permission_id
	INNER JOIN users
	ON users.id = upm.user_id
	WHERE users.id = $1;`

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions
	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (m *PermissionModel) AddForUser(userId int64, codes ...string) error {
	query := `
	INSERT INTO user_permissions
	SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2);`
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, userId, pq.Array(codes))
	return err
}