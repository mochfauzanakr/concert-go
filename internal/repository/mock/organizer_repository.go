package mock

import (
	"context"
	"sync"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

type organizerRepositoryMock struct {
	mu       sync.RWMutex
	profiles map[uuid.UUID]*entity.OrganizerProfile
}

func NewOrganizerRepositoryMock() repository.OrganizerRepository {
	return &organizerRepositoryMock{
		profiles: make(map[uuid.UUID]*entity.OrganizerProfile),
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
