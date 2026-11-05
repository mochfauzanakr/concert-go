package repository

import (
	"context"

	"concert-go/internal/domain/entity"

	"github.com/google/uuid"
)

// UserFilter defines options for querying user lists
type UserFilter struct {
	Search string
	RoleID *int
	Status string
	Limit  int
	Offset int
}

// UserRepository contract for user data persistence
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByProvider(ctx context.Context, provider, providerID string) (*entity.User, error)
	FindAll(ctx context.Context, filter UserFilter) ([]*entity.User, int, error)
}
