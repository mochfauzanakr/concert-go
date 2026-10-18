package mock

import (
	"context"
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
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	m.users[user.ID] = user
	return nil
}

func (m *userRepositoryMock) Update(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *userRepositoryMock) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.Email == email && !u.DeletedAt.Valid {
			return u, nil
		}
	}
	return nil, nil
}

func (m *userRepositoryMock) FindByProvider(ctx context.Context, provider, providerID string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.Provider == provider && u.ProviderID != nil && *u.ProviderID == providerID && !u.DeletedAt.Valid {
			return u, nil
		}
	}
	return nil, nil
}

func (m *userRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if u, exists := m.users[id]; exists && !u.DeletedAt.Valid {
		return u, nil
	}
	return nil, nil
}
