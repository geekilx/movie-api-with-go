package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Permissions []string

func (p Permissions) Include(code string) bool {
	for _, i := range p {
		if i == code {
			return true
		}
	}
	return false

}

type PermissionsModel struct {
	DB *sql.DB
}

func (m PermissionsModel) GetAllForUser(userID int64) (Permissions, error) {
	stmt := `
SELECT p.code
FROM permissions AS p
INNER JOIN users_permissions AS up ON up.permission_id = p.id
INNER JOIN users AS u ON up.user_id = u.id
WHERE u.id = $1
`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string
		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (m *PermissionsModel) AddForUser(userID int64, codes ...string) error {

	stmt := `INSERT INTO users_permissions
	SELECT $1, permissions.id FROM Permissions WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, userID, pq.Array(codes))
	return err

}
