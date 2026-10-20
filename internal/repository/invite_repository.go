package repository

import (
	"context"
	"time"

	"concert-go/internal/domain/entity"

	"github.com/google/uuid"
)

// InviteRepository contract for invite data persistence
type InviteRepository interface {
	Create(ctx context.Context, invite *entity.Invite) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Invite, error)
	FindPendingByEmail(ctx context.Context, email string) (*entity.Invite, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, acceptedAt *time.Time) error
}
