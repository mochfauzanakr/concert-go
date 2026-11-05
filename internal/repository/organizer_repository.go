package repository

import (
	"context"

	"concert-go/internal/domain/entity"

	"github.com/google/uuid"
)

type OrganizerRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.OrganizerProfile, error)
	Upsert(ctx context.Context, profile *entity.OrganizerProfile) error
}
