package repository

import (
	"context"
	"time"

	"concert-go/internal/domain/entity"

	"github.com/google/uuid"
)

type EventFilter struct {
	Search          string
	City            string
	Category        string
	Status          string
	ManagedByUserID *uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	SortBy          string
	SortOrder       string
	Limit           int
	Offset          int
}

type ArtistFilter struct {
	Search    string
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
	Create(ctx context.Context, event *entity.Event) error
	Update(ctx context.Context, event *entity.Event) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	FindArtistsByEventID(ctx context.Context, eventID uuid.UUID) ([]*entity.Artist, error)
	FindCoOrganizersByEventID(ctx context.Context, eventID uuid.UUID) ([]*EventCoOrganizer, error)
	FindAllArtists(ctx context.Context, filter ArtistFilter) ([]*entity.Artist, int, error)
	FindArtistByID(ctx context.Context, id uuid.UUID) (*entity.Artist, error)
	CreateArtist(ctx context.Context, artist *entity.Artist) error
	IsEventArtist(ctx context.Context, eventID, artistID uuid.UUID) (bool, error)
	AddEventArtist(ctx context.Context, eventID, artistID uuid.UUID) error
	RemoveEventArtist(ctx context.Context, eventID, artistID uuid.UUID) (bool, error)
}
