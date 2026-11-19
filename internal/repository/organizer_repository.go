package repository

import (
	"context"
	"time"

	"concert-go/internal/domain/entity"

	"github.com/google/uuid"
)

type OrganizerApplicationFilter struct {
	Search    string
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

type OrganizerRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.OrganizerProfile, error)
	FindByIDOrUserID(ctx context.Context, id uuid.UUID) (*entity.OrganizerProfile, error)
	Upsert(ctx context.Context, profile *entity.OrganizerProfile) error
	AddStaff(ctx context.Context, organizerID, userID uuid.UUID) error
	IsStaffOfOrganizer(ctx context.Context, organizerID, userID uuid.UUID) (bool, error)
	CreateApplication(ctx context.Context, app *entity.OrganizerApplication) error
	FindApplicationByID(ctx context.Context, id uuid.UUID) (*entity.OrganizerApplication, error)
	FindLatestApplicationByEmail(ctx context.Context, email string) (*entity.OrganizerApplication, error)
	FindAllApplications(ctx context.Context, filter OrganizerApplicationFilter) ([]*entity.OrganizerApplication, int, error)
	UpdateApplicationReview(ctx context.Context, app *entity.OrganizerApplication) error
}
