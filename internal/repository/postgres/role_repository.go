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
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		if role != nil {
			roles = append(roles, *role)
		}
	}

	return roles, rows.Err()
}

func (r *roleRepositoryImpl) FindByID(ctx context.Context, id int) (*entity.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles WHERE id = $1 LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanRole(row)
}

func (r *roleRepositoryImpl) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles WHERE UPPER(name) = UPPER($1) LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, name)
	return scanRole(row)
}

func scanRole(scanner interface{ Scan(dest ...interface{}) error }) (*entity.Role, error) {
	var role entity.Role
	var createdAt sql.NullTime
	var updatedAt sql.NullTime

	err := scanner.Scan(&role.ID, &role.Name, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if createdAt.Valid {
		role.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		role.UpdatedAt = updatedAt.Time
	}
	return &role, nil
}
