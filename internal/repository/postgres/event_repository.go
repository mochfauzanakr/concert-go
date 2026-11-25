package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/model"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

type eventRepositoryImpl struct {
	db *sql.DB
}

// NewEventRepository returns a pure SQL implementation of EventRepository.
func NewEventRepository(db *sql.DB) repository.EventRepository {
	return &eventRepositoryImpl{db: db}
}

func (r *eventRepositoryImpl) FindAll(ctx context.Context, filter repository.EventFilter) ([]*entity.Event, int, error) {
	where := []string{"e.deleted_at IS NULL"}
	var args []any
	argIdx := 1

	if filter.Status != "" {
		where = append(where, fmt.Sprintf("e.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.ManagedByUserID != nil {
		where = append(where, fmt.Sprintf(
			"(e.organizer_id = $%d OR EXISTS (SELECT 1 FROM event_organizers eo WHERE eo.event_id = e.id AND eo.user_id = $%d AND eo.role_in_event = $%d))",
			argIdx, argIdx, argIdx+1,
		))
		args = append(args, *filter.ManagedByUserID, entity.EventRoleCoOrganizer)
		argIdx += 2
	}
	if filter.City != "" {
		where = append(where, fmt.Sprintf("LOWER(e.city) = LOWER($%d)", argIdx))
		args = append(args, filter.City)
		argIdx++
	}
	if filter.Category != "" {
		where = append(where, fmt.Sprintf("LOWER(e.category) = LOWER($%d)", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf(
			"(LOWER(e.name) LIKE $%d OR EXISTS (SELECT 1 FROM event_artists ea JOIN artists a ON a.id = ea.artist_id WHERE ea.event_id = e.id AND LOWER(a.name) LIKE $%d))",
			argIdx, argIdx,
		))
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
		argIdx++
	}
	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.start_time >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.start_time <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events e WHERE %s", whereClause)
	if err := execer(ctx, r.db).QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		log.Printf("[SQL Error - Count Events] %v", err)
		return nil, 0, err
	}

	sortCol := model.ResolveSortColumn(filter.SortBy, map[string]string{
		"starttime":  "e.start_time",
		"start_time": "e.start_time",
		"name":       "e.name",
		"city":       "e.city",
		"category":   "e.category",
		"createdat":  "e.created_at",
		"created_at": "e.created_at",
	}, "e.created_at")
	sortOrder := model.ResolveSortOrder(filter.SortOrder)

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	dataQuery := fmt.Sprintf(`
		SELECT e.id, e.organizer_id, e.name, e.description, e.banner_url, e.category, e.city, e.location, e.start_time, e.end_time, e.status, e.inventory_mode, e.created_at, e.updated_at, e.deleted_at
		FROM events e
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, sortOrder, argIdx, argIdx+1)

	rows, err := execer(ctx, r.db).QueryContext(ctx, dataQuery, append(args, limit, filter.Offset)...)
	if err != nil {
		log.Printf("[SQL Error - Query Events] %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	events := make([]*entity.Event, 0)
	for rows.Next() {
		ev, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		if ev != nil {
			events = append(events, ev)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("[SQL Error - Events Rows] %v", err)
		return nil, 0, err
	}

	return events, total, nil
}

func (r *eventRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	query := `
		SELECT id, organizer_id, name, description, banner_url, category, city, location, start_time, end_time, status, inventory_mode, created_at, updated_at, deleted_at
		FROM events
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	row := execer(ctx, r.db).QueryRowContext(ctx, query, id)
	return scanEvent(row)
}

func (r *eventRepositoryImpl) Create(ctx context.Context, event *entity.Event) error {
	now := time.Now()
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Status == "" {
		event.Status = entity.EventStatusDraft
	}
	if event.InventoryMode == "" {
		event.InventoryMode = entity.InventoryModeOnlineOnly
	}
	event.CreatedAt = now
	event.UpdatedAt = now

	query := `
		INSERT INTO events (id, organizer_id, name, description, banner_url, category, city, location, start_time, end_time, status, inventory_mode, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := execer(ctx, r.db).ExecContext(ctx, query,
		event.ID,
		event.OrganizerID,
		event.Name,
		event.Description,
		event.BannerURL,
		event.Category,
		event.City,
		event.Location,
		event.StartTime,
		event.EndTime,
		event.Status,
		event.InventoryMode,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		log.Printf("[SQL Error - Event Create] %v", err)
	}
	return err
}

func (r *eventRepositoryImpl) Update(ctx context.Context, event *entity.Event) error {
	event.UpdatedAt = time.Now()
	query := `
		UPDATE events
		SET name = $1, description = $2, banner_url = $3, category = $4, city = $5, location = $6, start_time = $7, end_time = $8, inventory_mode = $9, updated_at = $10
		WHERE id = $11 AND deleted_at IS NULL
	`
	_, err := execer(ctx, r.db).ExecContext(ctx, query,
		event.Name,
		event.Description,
		event.BannerURL,
		event.Category,
		event.City,
		event.Location,
		event.StartTime,
		event.EndTime,
		event.InventoryMode,
		event.UpdatedAt,
		event.ID,
	)
	if err != nil {
		log.Printf("[SQL Error - Event Update] %v", err)
	}
	return err
}

func (r *eventRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE events SET status = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
	_, err := execer(ctx, r.db).ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		log.Printf("[SQL Error - Event UpdateStatus] %v", err)
	}
	return err
}

func (r *eventRepositoryImpl) FindArtistsByEventID(ctx context.Context, eventID uuid.UUID) ([]*entity.Artist, error) {
	query := `
		SELECT a.id, a.name, a.bio, a.avatar_url, a.created_at, a.updated_at
		FROM event_artists ea
		JOIN artists a ON a.id = ea.artist_id
		WHERE ea.event_id = $1
		ORDER BY a.name ASC
	`
	rows, err := execer(ctx, r.db).QueryContext(ctx, query, eventID)
	if err != nil {
		log.Printf("[SQL Error - Query Event Artists] %v", err)
		return nil, err
	}
	defer rows.Close()

	artists := make([]*entity.Artist, 0)
	for rows.Next() {
		var a entity.Artist
		var bio, avatarURL sql.NullString
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&a.ID, &a.Name, &bio, &avatarURL, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if bio.Valid {
			a.Bio = &bio.String
		}
		if avatarURL.Valid {
			a.AvatarURL = &avatarURL.String
		}
		if createdAt.Valid {
			a.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			a.UpdatedAt = updatedAt.Time
		}
		artists = append(artists, &a)
	}
	return artists, rows.Err()
}

func (r *eventRepositoryImpl) FindCoOrganizersByEventID(ctx context.Context, eventID uuid.UUID) ([]*repository.EventCoOrganizer, error) {
	query := `
		SELECT eo.user_id, u.name, COALESCE(op.org_name, u.name) AS org_name, op.logo_url, eo.assigned_at
		FROM event_organizers eo
		JOIN users u ON u.id = eo.user_id AND u.deleted_at IS NULL
		LEFT JOIN organizer_profiles op ON op.user_id = eo.user_id
		WHERE eo.event_id = $1 AND eo.role_in_event = $2
		ORDER BY eo.assigned_at ASC
	`
	rows, err := execer(ctx, r.db).QueryContext(ctx, query, eventID, entity.EventRoleCoOrganizer)
	if err != nil {
		log.Printf("[SQL Error - Query Event Co-Organizers] %v", err)
		return nil, err
	}
	defer rows.Close()

	coOrganizers := make([]*repository.EventCoOrganizer, 0)
	for rows.Next() {
		var c repository.EventCoOrganizer
		var logoURL sql.NullString

		if err := rows.Scan(&c.UserID, &c.Name, &c.OrgName, &logoURL, &c.AssignedAt); err != nil {
			return nil, err
		}
		if logoURL.Valid {
			c.LogoURL = &logoURL.String
		}
		coOrganizers = append(coOrganizers, &c)
	}
	return coOrganizers, rows.Err()
}

func scanEvent(row interface{ Scan(dest ...any) error }) (*entity.Event, error) {
	var ev entity.Event
	var bannerURL sql.NullString
	var createdAt, updatedAt, deletedAt sql.NullTime

	err := row.Scan(
		&ev.ID,
		&ev.OrganizerID,
		&ev.Name,
		&ev.Description,
		&bannerURL,
		&ev.Category,
		&ev.City,
		&ev.Location,
		&ev.StartTime,
		&ev.EndTime,
		&ev.Status,
		&ev.InventoryMode,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Printf("[SQL Error - Scan Event] %v", err)
		return nil, err
	}

	if bannerURL.Valid {
		ev.BannerURL = &bannerURL.String
	}
	if createdAt.Valid {
		ev.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		ev.UpdatedAt = updatedAt.Time
	}
	if deletedAt.Valid {
		ev.DeletedAt.Time = deletedAt.Time
		ev.DeletedAt.Valid = true
	}

	return &ev, nil
}
