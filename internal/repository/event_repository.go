package repository

import (
	"context"
	"time"

	"concert-go/internal/domain/entity"

	"github.com/google/uuid"
)

type EventFilter struct {
	Search    string
	City      string
	Category  string
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

type EventCoOrganizer struct {
	UserID     uuid.UUID
	Name       string
	OrgName    string
	LogoURL    *string
	AssignedAt time.Time
}

type EventRepository interface {
	FindAll(ctx context.Context, filter EventFilter) ([]*entity.Event, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Event, error)
	FindArtistsByEventID(ctx context.Context, eventID uuid.UUID) ([]*entity.Artist, error)
	FindCoOrganizersByEventID(ctx context.Context, eventID uuid.UUID) ([]*EventCoOrganizer, error)
}
