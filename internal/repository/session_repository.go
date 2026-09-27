package repository

import (
	"context"

	"concert-go/internal/domain/entity"
)

// SessionRepository contract for stateful session persistence
type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error)
	Revoke(ctx context.Context, tokenHash string) error
}
