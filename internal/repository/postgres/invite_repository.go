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

type inviteRepositoryImpl struct {
	db *sql.DB
}

// NewInviteRepository creates a pure SQL implementation of InviteRepository
func NewInviteRepository(db *sql.DB) repository.InviteRepository {
	return &inviteRepositoryImpl{db: db}
}

func (r *inviteRepositoryImpl) Create(ctx context.Context, invite *entity.Invite) error {
	query := `
		INSERT INTO invites (id, email, role_id, invited_by, token_hash, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now()
	if invite.ID == uuid.Nil {
		invite.ID = uuid.New()
	}
	if invite.Status == "" {
		invite.Status = entity.InviteStatusPending
	}
	invite.CreatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		invite.ID,
		invite.Email,
		invite.RoleID,
		invite.InvitedBy,
		invite.TokenHash,
		invite.Status,
		invite.ExpiresAt,
		invite.CreatedAt,
	)
	if err != nil {
		log.Printf("[SQL Error - Invite Create] %v", err)
	}
	return err
}

func (r *inviteRepositoryImpl) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Invite, error) {
	query := `
		SELECT id, email, role_id, invited_by, token_hash, status, expires_at, created_at, accepted_at
		FROM invites
		WHERE token_hash = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, tokenHash)
	return scanInvite(row)
}

func (r *inviteRepositoryImpl) FindPendingByEmail(ctx context.Context, email string) (*entity.Invite, error) {
	query := `
		SELECT id, email, role_id, invited_by, token_hash, status, expires_at, created_at, accepted_at
		FROM invites
		WHERE LOWER(email) = LOWER($1) AND status = $2 AND expires_at > $3
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, email, entity.InviteStatusPending, time.Now())
	return scanInvite(row)
}

func (r *inviteRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status string, acceptedAt *time.Time) error {
	query := `
		UPDATE invites
		SET status = $1, accepted_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, status, acceptedAt, id)
	return err
}

func scanInvite(row interface{ Scan(dest ...interface{}) error }) (*entity.Invite, error) {
	var inv entity.Invite
	var acceptedAt sql.NullTime

	err := row.Scan(
		&inv.ID,
		&inv.Email,
		&inv.RoleID,
		&inv.InvitedBy,
		&inv.TokenHash,
		&inv.Status,
		&inv.ExpiresAt,
		&inv.CreatedAt,
		&acceptedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Printf("[SQL Error - Scan Invite] %v", err)
		return nil, err
	}

	if acceptedAt.Valid {
		inv.AcceptedAt = &acceptedAt.Time
	}

	return &inv, nil
}
