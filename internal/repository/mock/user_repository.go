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

type userRepositoryMock struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*entity.User
}

// NewUserRepositoryMock creates in-memory thread-safe mock user repository
func NewUserRepositoryMock() repository.UserRepository {
	return &userRepositoryMock{
		users: make(map[uuid.UUID]*entity.User),
	}
}

func (m *userRepositoryMock) Create(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.Status == "" {
		user.Status = entity.UserStatusActive
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	prev, existed := m.users[user.ID]
	copyUser := *user
	m.users[user.ID] = &copyUser
	userID := user.ID
	registerRollback(ctx, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if existed {
			m.users[userID] = prev
		} else {
			delete(m.users, userID)
		}
	})
	return nil
}

func (m *userRepositoryMock) Update(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user.UpdatedAt = time.Now()
	prev, existed := m.users[user.ID]
	copyUser := *user
	m.users[user.ID] = &copyUser
	userID := user.ID
	registerRollback(ctx, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if existed {
			m.users[userID] = prev
		} else {
			delete(m.users, userID)
		}
	})
	return nil
}

func (m *userRepositoryMock) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.Email == email && !u.DeletedAt.Valid {
			copyUser := *u
			return &copyUser, nil
		}
	}
	return nil, nil
}

func (m *userRepositoryMock) FindByProvider(ctx context.Context, provider, providerID string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.Provider == provider && u.ProviderID != nil && *u.ProviderID == providerID && !u.DeletedAt.Valid {
			copyUser := *u
			return &copyUser, nil
		}
	}
	return nil, nil
}

func (m *userRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if u, exists := m.users[id]; exists && !u.DeletedAt.Valid {
		copyUser := *u
		return &copyUser, nil
	}
	return nil, nil
}

func (m *userRepositoryMock) FindAll(ctx context.Context, filter repository.UserFilter) ([]*entity.User, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []*entity.User
	searchLower := strings.ToLower(strings.TrimSpace(filter.Search))

	for _, u := range m.users {
		if u.DeletedAt.Valid {
			continue
		}
		if filter.OrganizerID != nil {
			if !isMockOrganizerStaff(*filter.OrganizerID, u.ID) {
				continue
			}
		}
		if searchLower != "" {
			nameMatch := strings.Contains(strings.ToLower(u.Name), searchLower)
			emailMatch := strings.Contains(strings.ToLower(u.Email), searchLower)
			if !nameMatch && !emailMatch {
				continue
			}
		}
		if filter.RoleID != nil {
			if u.RoleID == nil || *u.RoleID != *filter.RoleID {
				continue
			}
		}
		if filter.Status != "" && u.Status != filter.Status {
			continue
		}
		if filter.StartDate != nil && u.CreatedAt.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && u.CreatedAt.After(*filter.EndDate) {
			continue
		}
		copyUser := *u
		filtered = append(filtered, &copyUser)
	}

	asc := strings.EqualFold(strings.TrimSpace(filter.SortOrder), "ASC")
	sortBy := strings.ToLower(strings.TrimSpace(filter.SortBy))
	sort.Slice(filtered, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
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
	start := filter.Offset
	if start > total {
		start = total
	}
	end := start + filter.Limit
	if filter.Limit <= 0 || end > total {
		end = total
	}

	return filtered[start:end], total, nil
}
