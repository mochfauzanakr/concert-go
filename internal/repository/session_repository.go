package repository

import (
	"context"

	"concert-go/internal/domain/entity"

	"github.com/google/uuid"
)

// SessionRepository contract for stateful session persistence
type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
}
