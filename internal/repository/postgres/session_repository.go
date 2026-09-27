package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

type sessionRepositoryImpl struct {
	db *sql.DB
}

// NewSessionRepository returns pure SQL implementation of SessionRepository
func NewSessionRepository(db *sql.DB) repository.SessionRepository {
	return &sessionRepositoryImpl{db: db}
}

func (r *sessionRepositoryImpl) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token_hash, ip_address, user_agent, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	session.CreatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.CreatedAt,
	)
	return err
}

func (r *sessionRepositoryImpl) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error) {
	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, revoked_at, created_at
		FROM sessions
		WHERE token_hash = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, tokenHash)

	var s entity.Session
	var revokedAt sql.NullTime

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.TokenHash,
		&s.IPAddress,
		&s.UserAgent,
		&s.ExpiresAt,
		&revokedAt,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if revokedAt.Valid {
		s.RevokedAt = &revokedAt.Time
	}
	return &s, nil
}

func (r *sessionRepositoryImpl) Revoke(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE sessions
		SET revoked_at = $1
		WHERE token_hash = $2 AND revoked_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), tokenHash)
	return err
}
