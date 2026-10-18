package postgres

import (
	"context"
	"database/sql"
	"errors"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"
)

type roleRepositoryImpl struct {
	db *sql.DB
}

// NewRoleRepository returns a pure SQL implementation of RoleRepository
func NewRoleRepository(db *sql.DB) repository.RoleRepository {
	return &roleRepositoryImpl{db: db}
}

func (r *roleRepositoryImpl) FindAll(ctx context.Context) ([]entity.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []entity.Role
	for rows.Next() {
		var role entity.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (r *roleRepositoryImpl) FindByID(ctx context.Context, id int) (*entity.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles WHERE id = $1 LIMIT 1`
	var role entity.Role
	err := r.db.QueryRowContext(ctx, query, id).Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepositoryImpl) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles WHERE UPPER(name) = UPPER($1) LIMIT 1`
	var role entity.Role
	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}
