package repository

import (
	"context"

	"concert-go/internal/domain/entity"
)

type RoleRepository interface {
	FindAll(ctx context.Context) ([]entity.Role, error)
	FindByID(ctx context.Context, id int) (*entity.Role, error)
	FindByName(ctx context.Context, name string) (*entity.Role, error)
}
