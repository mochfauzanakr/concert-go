package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

type organizerRepositoryImpl struct {
	db *sql.DB
}

func NewOrganizerRepository(db *sql.DB) repository.OrganizerRepository {
	return &organizerRepositoryImpl{db: db}
}

func (r *organizerRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.OrganizerProfile, error) {
	query := `
		SELECT id, user_id, org_name, logo_url, description, phone, social_media, created_at, updated_at
		FROM organizer_profiles
		WHERE user_id = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, userID)

	var p entity.OrganizerProfile
	var logoURL, description, phone, socialMedia sql.NullString

	err := row.Scan(
		&p.ID,
		&p.UserID,
		&p.OrgName,
		&logoURL,
		&description,
		&phone,
		&socialMedia,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if logoURL.Valid {
		p.LogoURL = &logoURL.String
	}
	if description.Valid {
		p.Description = &description.String
	}
	if phone.Valid {
		p.Phone = &phone.String
	}
	if socialMedia.Valid {
		p.SocialMedia = &socialMedia.String
	}

	return &p, nil
}

func (r *organizerRepositoryImpl) Upsert(ctx context.Context, profile *entity.OrganizerProfile) error {
	now := time.Now()
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}

	query := `
		INSERT INTO organizer_profiles (id, user_id, org_name, logo_url, description, phone, social_media, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			org_name = EXCLUDED.org_name,
			logo_url = COALESCE(EXCLUDED.logo_url, organizer_profiles.logo_url),
			description = EXCLUDED.description,
			phone = EXCLUDED.phone,
			social_media = EXCLUDED.social_media,
			updated_at = EXCLUDED.updated_at
		RETURNING id, logo_url, created_at, updated_at
	`

	var logoURL sql.NullString
	err := r.db.QueryRowContext(ctx, query,
		profile.ID,
		profile.UserID,
		profile.OrgName,
		profile.LogoURL,
		profile.Description,
		profile.Phone,
		profile.SocialMedia,
		now,
		now,
	).Scan(&profile.ID, &logoURL, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		return err
	}

	if logoURL.Valid {
		profile.LogoURL = &logoURL.String
	}

	return nil
}
