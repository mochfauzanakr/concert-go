package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

type userRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository returns pure SQL implementation of UserRepository
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, role_id, name, email, password_hash, provider, provider_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	now := time.Now()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.Provider == "" {
		user.Provider = "email"
	}
	if user.Status == "" {
		user.Status = entity.UserStatusActive
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.RoleID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Provider,
		user.ProviderID,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		log.Printf("[SQL Error - User Create] %v", err)
	}
	return err
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET role_id = $1, name = $2, email = $3, password_hash = $4, provider = $5, provider_id = $6, status = $7, updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
	`
	if user.Status == "" {
		user.Status = entity.UserStatusActive
	}
	user.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		user.RoleID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Provider,
		user.ProviderID,
		user.Status,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		log.Printf("[SQL Error - User Update] %v", err)
	}
	return err
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, role_id, name, email, password_hash, provider, provider_id, status, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, email)
	return scanUser(row)
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, role_id, name, email, password_hash, provider, provider_id, status, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanUser(row)
}

func (r *userRepositoryImpl) FindByProvider(ctx context.Context, provider, providerID string) (*entity.User, error) {
	query := `
		SELECT id, role_id, name, email, password_hash, provider, provider_id, status, created_at, updated_at, deleted_at
		FROM users
		WHERE provider = $1 AND provider_id = $2 AND deleted_at IS NULL
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, provider, providerID)
	return scanUser(row)
}

func scanUser(row interface{ Scan(dest ...interface{}) error }) (*entity.User, error) {
	var u entity.User
	var deletedAt sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var roleID sql.NullInt64
	var passwordHash sql.NullString
	var providerID sql.NullString
	var status sql.NullString

	err := row.Scan(
		&u.ID,
		&roleID,
		&u.Name,
		&u.Email,
		&passwordHash,
		&u.Provider,
		&providerID,
		&status,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Printf("[SQL Error - Scan User] %v", err)
		return nil, err
	}

	if roleID.Valid {
		id := int(roleID.Int64)
		u.RoleID = &id
	}
	if passwordHash.Valid {
		u.PasswordHash = &passwordHash.String
	}
	if providerID.Valid {
		u.ProviderID = &providerID.String
	}
	if status.Valid {
		u.Status = status.String
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	return &u, nil
}
