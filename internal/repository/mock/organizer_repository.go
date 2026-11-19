package mock

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

var (
	staffRegistryMu sync.RWMutex
	staffRegistry   = make(map[uuid.UUID]map[uuid.UUID]struct{})
)

func isMockOrganizerStaff(organizerID, userID uuid.UUID) bool {
	staffRegistryMu.RLock()
	defer staffRegistryMu.RUnlock()
	if set, ok := staffRegistry[organizerID]; ok {
		_, exists := set[userID]
		return exists
	}
	return false
}

type organizerRepositoryMock struct {
	mu           sync.RWMutex
	profiles     map[uuid.UUID]*entity.OrganizerProfile
	applications map[uuid.UUID]*entity.OrganizerApplication
}

func NewOrganizerRepositoryMock() repository.OrganizerRepository {
	return &organizerRepositoryMock{
		profiles:     make(map[uuid.UUID]*entity.OrganizerProfile),
		applications: make(map[uuid.UUID]*entity.OrganizerApplication),
	}
}

func (m *organizerRepositoryMock) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.OrganizerProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, exists := m.profiles[userID]; exists {
		copyProfile := *p
		return &copyProfile, nil
	}
	return nil, nil
}

func (m *organizerRepositoryMock) FindByIDOrUserID(ctx context.Context, id uuid.UUID) (*entity.OrganizerProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, exists := m.profiles[id]; exists {
		copyProfile := *p
		return &copyProfile, nil
	}
	for _, p := range m.profiles {
		if p.ID == id {
			copyProfile := *p
			return &copyProfile, nil
		}
	}
	return nil, nil
}

func (m *organizerRepositoryMock) Upsert(ctx context.Context, profile *entity.OrganizerProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if existing, exists := m.profiles[profile.UserID]; exists {
		profile.ID = existing.ID
		profile.CreatedAt = existing.CreatedAt
		profile.UpdatedAt = now
		if profile.LogoURL == nil {
			profile.LogoURL = existing.LogoURL
		}
	} else {
		if profile.ID == uuid.Nil {
			profile.ID = uuid.New()
		}
		profile.CreatedAt = now
		profile.UpdatedAt = now
	}

	copyProfile := *profile
	m.profiles[profile.UserID] = &copyProfile
	return nil
}

func (m *organizerRepositoryMock) AddStaff(ctx context.Context, organizerID, userID uuid.UUID) error {
	staffRegistryMu.Lock()
	defer staffRegistryMu.Unlock()

	if staffRegistry[organizerID] == nil {
		staffRegistry[organizerID] = make(map[uuid.UUID]struct{})
	}
	staffRegistry[organizerID][userID] = struct{}{}
	return nil
}

func (m *organizerRepositoryMock) IsStaffOfOrganizer(ctx context.Context, organizerID, userID uuid.UUID) (bool, error) {
	return isMockOrganizerStaff(organizerID, userID), nil
}

func (m *organizerRepositoryMock) CreateApplication(ctx context.Context, app *entity.OrganizerApplication) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	if app.CreatedAt.IsZero() {
		app.CreatedAt = time.Now()
	}
	if app.Status == "" {
		app.Status = entity.ApplicationStatusPending
	}

	copyApp := *app
	m.applications[app.ID] = &copyApp
	return nil
}

func (m *organizerRepositoryMock) FindApplicationByID(ctx context.Context, id uuid.UUID) (*entity.OrganizerApplication, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if app, exists := m.applications[id]; exists {
		copyApp := *app
		return &copyApp, nil
	}
	return nil, nil
}

func (m *organizerRepositoryMock) FindLatestApplicationByEmail(ctx context.Context, email string) (*entity.OrganizerApplication, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latest *entity.OrganizerApplication
	for _, app := range m.applications {
		if strings.EqualFold(app.Email, email) {
			if latest == nil || !app.CreatedAt.Before(latest.CreatedAt) {
				copyApp := *app
				latest = &copyApp
			}
		}
	}
	return latest, nil
}

func (m *organizerRepositoryMock) FindAllApplications(ctx context.Context, filter repository.OrganizerApplicationFilter) ([]*entity.OrganizerApplication, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	searchLower := strings.ToLower(strings.TrimSpace(filter.Search))
	var filtered []*entity.OrganizerApplication

	for _, app := range m.applications {
		if filter.Status != "" && !strings.EqualFold(app.Status, filter.Status) {
			continue
		}
		if filter.StartDate != nil && app.CreatedAt.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && app.CreatedAt.After(*filter.EndDate) {
			continue
		}
		if searchLower != "" {
			matchPic := strings.Contains(strings.ToLower(app.PicName), searchLower)
			matchEmail := strings.Contains(strings.ToLower(app.Email), searchLower)
			matchOrg := strings.Contains(strings.ToLower(app.OrgName), searchLower)
			if !matchPic && !matchEmail && !matchOrg {
				continue
			}
		}
		copyApp := *app
		filtered = append(filtered, &copyApp)
	}

	asc := strings.EqualFold(strings.TrimSpace(filter.SortOrder), "ASC")
	sortBy := strings.ToLower(strings.TrimSpace(filter.SortBy))
	sort.Slice(filtered, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "picname", "pic_name":
			less = strings.ToLower(filtered[i].PicName) < strings.ToLower(filtered[j].PicName)
		case "orgname", "org_name":
			less = strings.ToLower(filtered[i].OrgName) < strings.ToLower(filtered[j].OrgName)
		case "email":
			less = strings.ToLower(filtered[i].Email) < strings.ToLower(filtered[j].Email)
		default:
			less = filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
		if asc {
			return less
		}
		return !less
	})

	total := len(filtered)
	offset := filter.Offset
	if offset >= total {
		return []*entity.OrganizerApplication{}, total, nil
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

func (m *organizerRepositoryMock) UpdateApplicationReview(ctx context.Context, app *entity.OrganizerApplication) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.applications[app.ID]
	if !exists {
		return nil
	}
	existing.Status = app.Status
	existing.RejectionReason = app.RejectionReason
	existing.ReviewedBy = app.ReviewedBy
	existing.ReviewedAt = app.ReviewedAt
	return nil
}
