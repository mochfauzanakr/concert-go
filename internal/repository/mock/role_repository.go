package mock

import (
	"context"
	"strings"
	"sync"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"
)

type roleRepositoryMock struct {
	mu    sync.RWMutex
	roles map[int]entity.Role
}

// NewRoleRepositoryMock creates in-memory thread-safe mock role repository pre-populated with standard roles
func NewRoleRepositoryMock() repository.RoleRepository {
	now := time.Now()
	m := &roleRepositoryMock{
		roles: make(map[int]entity.Role),
	}

	defaultRoles := []string{"SUPER_ADMIN", "ADMIN", "ORGANIZER", "USER", "STAFF"}
	for i, name := range defaultRoles {
		id := i + 1
		m.roles[id] = entity.Role{
			ID:        id,
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return m
}

func (m *roleRepositoryMock) FindAll(ctx context.Context) ([]entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []entity.Role
	for i := 1; i <= len(m.roles); i++ {
		if r, ok := m.roles[i]; ok {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *roleRepositoryMock) FindByID(ctx context.Context, id int) (*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if r, ok := m.roles[id]; ok {
		roleCopy := r
		return &roleCopy, nil
	}
	return nil, nil
}

func (m *roleRepositoryMock) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, r := range m.roles {
		if strings.EqualFold(r.Name, name) {
			roleCopy := r
			return &roleCopy, nil
		}
	}
	return nil, nil
}
