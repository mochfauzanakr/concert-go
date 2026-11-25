package mock

import (
	"context"
	"strings"
	"sync"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

type inviteRepositoryMock struct {
	mu      sync.RWMutex
	invites map[uuid.UUID]*entity.Invite
}

// NewInviteRepositoryMock creates in-memory thread-safe mock invite repository
func NewInviteRepositoryMock() repository.InviteRepository {
	return &inviteRepositoryMock{
		invites: make(map[uuid.UUID]*entity.Invite),
	}
}

func (m *inviteRepositoryMock) Create(ctx context.Context, invite *entity.Invite) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if invite.ID == uuid.Nil {
		invite.ID = uuid.New()
	}
	now := time.Now()
	invite.CreatedAt = now
	if invite.Status == "" {
		invite.Status = entity.InviteStatusPending
	}

	prev, existed := m.invites[invite.ID]
	copyInv := *invite
	m.invites[invite.ID] = &copyInv
	inviteID := invite.ID
	registerRollback(ctx, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if existed {
			m.invites[inviteID] = prev
		} else {
			delete(m.invites, inviteID)
		}
	})
	return nil
}

func (m *inviteRepositoryMock) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Invite, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, inv := range m.invites {
		if inv.TokenHash == tokenHash {
			copyInv := *inv
			return &copyInv, nil
		}
	}
	return nil, nil
}

func (m *inviteRepositoryMock) FindPendingByEmail(ctx context.Context, email string) (*entity.Invite, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	for _, inv := range m.invites {
		if strings.EqualFold(inv.Email, email) && inv.Status == entity.InviteStatusPending && inv.ExpiresAt.After(now) {
			copyInv := *inv
			return &copyInv, nil
		}
	}
	return nil, nil
}

func (m *inviteRepositoryMock) UpdateStatus(ctx context.Context, id uuid.UUID, status string, acceptedAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inv, exists := m.invites[id]; exists {
		prevStatus := inv.Status
		prevAcceptedAt := inv.AcceptedAt
		inv.Status = status
		inv.AcceptedAt = acceptedAt
		registerRollback(ctx, func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			if cur, ok := m.invites[id]; ok {
				cur.Status = prevStatus
				cur.AcceptedAt = prevAcceptedAt
			}
		})
	}
	return nil
}
