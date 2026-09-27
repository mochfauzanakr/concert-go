package mock

import (
	"concert-go/internal/repository"
	"sync"
)

type helloRepositoryMock struct {
	mu sync.RWMutex
}

// NewHelloRepositoryMock implements in-memory mock repository
func NewHelloRepositoryMock() repository.HelloRepository {
	return &helloRepositoryMock{}
}

func (m *helloRepositoryMock) PingDatabase() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return nil
}
