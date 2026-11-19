package repository

import (
	"context"

	"concert-go/internal/domain/entity"

	"github.com/google/uuid"
)

type UserFilter struct {
	Search      string
	RoleID      *int
	Status      string
	OrganizerID *uuid.UUID
	Limit       int
	Offset      int
}

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByProvider(ctx context.Context, provider, providerID string) (*entity.User, error)
	FindAll(ctx context.Context, filter UserFilter) ([]*entity.User, int, error)
}
