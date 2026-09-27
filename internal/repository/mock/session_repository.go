package mock

import (
	"context"
	"sync"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

type sessionRepositoryMock struct {
	mu       sync.RWMutex
	sessions map[string]*entity.Session
}

// NewSessionRepositoryMock creates in-memory thread-safe mock session repository
func NewSessionRepositoryMock() repository.SessionRepository {
	return &sessionRepositoryMock{
		sessions: make(map[string]*entity.Session),
	}
}

func (m *sessionRepositoryMock) Create(ctx context.Context, session *entity.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	session.CreatedAt = time.Now()

	m.sessions[session.TokenHash] = session
	return nil
}

func (m *sessionRepositoryMock) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if s, exists := m.sessions[tokenHash]; exists {
		return s, nil
	}
	return nil, nil
}

func (m *sessionRepositoryMock) Revoke(ctx context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, exists := m.sessions[tokenHash]; exists {
		now := time.Now()
		s.RevokedAt = &now
	}
	return nil
}
