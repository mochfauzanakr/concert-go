package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/model"
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
	return r.scanRow(r.db.QueryRowContext(ctx, query, userID))
}

func (r *organizerRepositoryImpl) FindByIDOrUserID(ctx context.Context, id uuid.UUID) (*entity.OrganizerProfile, error) {
	query := `
		SELECT id, user_id, org_name, logo_url, description, phone, social_media, created_at, updated_at
		FROM organizer_profiles
		WHERE id = $1 OR user_id = $1
		LIMIT 1
	`
	return r.scanRow(r.db.QueryRowContext(ctx, query, id))
}

func (r *organizerRepositoryImpl) scanRow(row *sql.Row) (*entity.OrganizerProfile, error) {
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

func (r *organizerRepositoryImpl) AddStaff(ctx context.Context, organizerID, userID uuid.UUID) error {
	query := `
		INSERT INTO organizer_staffs (id, organizer_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (organizer_id, user_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New(), organizerID, userID, time.Now())
	return err
}

func (r *organizerRepositoryImpl) IsStaffOfOrganizer(ctx context.Context, organizerID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizer_staffs WHERE organizer_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, organizerID, userID).Scan(&exists)
	return exists, err
}

type applicationScanner interface {
	Scan(dest ...interface{}) error
}

func scanApplication(s applicationScanner) (*entity.OrganizerApplication, error) {
	var a entity.OrganizerApplication
	var socialMedia, rejectionReason sql.NullString
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime

	err := s.Scan(
		&a.ID,
		&a.PicName,
		&a.Email,
		&a.Phone,
		&a.OrgName,
		&socialMedia,
		&a.Description,
		&a.Status,
		&rejectionReason,
		&reviewedBy,
		&reviewedAt,
		&a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if socialMedia.Valid {
		a.SocialMedia = &socialMedia.String
	}
	if rejectionReason.Valid {
		a.RejectionReason = &rejectionReason.String
	}
	if reviewedBy.Valid {
		if id, err := uuid.Parse(reviewedBy.String); err == nil {
			a.ReviewedBy = &id
		}
	}
	if reviewedAt.Valid {
		a.ReviewedAt = &reviewedAt.Time
	}

	return &a, nil
}

func (r *organizerRepositoryImpl) CreateApplication(ctx context.Context, app *entity.OrganizerApplication) error {
	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	if app.CreatedAt.IsZero() {
		app.CreatedAt = time.Now()
	}
	if app.Status == "" {
		app.Status = entity.ApplicationStatusPending
	}

	query := `
		INSERT INTO organizer_applications (id, pic_name, email, phone, org_name, social_media, description, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		app.ID,
		app.PicName,
		app.Email,
		app.Phone,
		app.OrgName,
		app.SocialMedia,
		app.Description,
		app.Status,
		app.CreatedAt,
	)
	return err
}

func (r *organizerRepositoryImpl) FindApplicationByID(ctx context.Context, id uuid.UUID) (*entity.OrganizerApplication, error) {
	query := `
		SELECT id, pic_name, email, phone, org_name, social_media, description, status, rejection_reason, reviewed_by, reviewed_at, created_at
		FROM organizer_applications
		WHERE id = $1
		LIMIT 1
	`
	return scanApplication(r.db.QueryRowContext(ctx, query, id))
}

func (r *organizerRepositoryImpl) FindLatestApplicationByEmail(ctx context.Context, email string) (*entity.OrganizerApplication, error) {
	query := `
		SELECT id, pic_name, email, phone, org_name, social_media, description, status, rejection_reason, reviewed_by, reviewed_at, created_at
		FROM organizer_applications
		WHERE LOWER(email) = LOWER($1)
		ORDER BY created_at DESC
		LIMIT 1
	`
	return scanApplication(r.db.QueryRowContext(ctx, query, email))
}

func (r *organizerRepositoryImpl) FindAllApplications(ctx context.Context, filter repository.OrganizerApplicationFilter) ([]*entity.OrganizerApplication, int, error) {
	where := []string{"1=1"}
	var args []interface{}
	argIdx := 1

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(LOWER(pic_name) LIKE $%d OR LOWER(email) LIKE $%d OR LOWER(org_name) LIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
		argIdx++
	}

	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM organizer_applications WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	sortCol := model.ResolveSortColumn(filter.SortBy, map[string]string{
		"createdat":  "created_at",
		"created_at": "created_at",
		"picname":    "pic_name",
		"pic_name":   "pic_name",
		"orgname":    "org_name",
		"org_name":   "org_name",
		"email":      "email",
		"status":     "status",
	}, "created_at")
	sortOrder := model.ResolveSortOrder(filter.SortOrder)

	dataQuery := fmt.Sprintf(`
		SELECT id, pic_name, email, phone, org_name, social_media, description, status, rejection_reason, reviewed_by, reviewed_at, created_at
		FROM organizer_applications
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, sortOrder, argIdx, argIdx+1)

	rows, err := r.db.QueryContext(ctx, dataQuery, append(args, limit, filter.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	apps := make([]*entity.OrganizerApplication, 0)
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, 0, err
		}
		if a != nil {
			apps = append(apps, a)
		}
	}
	return apps, total, rows.Err()
}

func (r *organizerRepositoryImpl) UpdateApplicationReview(ctx context.Context, app *entity.OrganizerApplication) error {
	query := `
		UPDATE organizer_applications
		SET status = $1, rejection_reason = $2, reviewed_by = $3, reviewed_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, app.Status, app.RejectionReason, app.ReviewedBy, app.ReviewedAt, app.ID)
	return err
}
