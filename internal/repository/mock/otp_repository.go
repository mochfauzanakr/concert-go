package mock

import (
	"context"
	"strings"
	"sync"
	"time"

	"concert-go/internal/repository"
)

type pendingRegItem struct {
	reg       repository.PendingRegistration
	expiresAt time.Time
}

type resetOtpItem struct {
	otp       string
	expiresAt time.Time
}

type otpRepositoryMock struct {
	mu        sync.RWMutex
	data      map[string]pendingRegItem
	resetData map[string]resetOtpItem
}

// NewOTPRepositoryMock creates an in-memory thread-safe mock OTP repository
func NewOTPRepositoryMock() repository.OTPRepository {
	return &otpRepositoryMock{
		data:      make(map[string]pendingRegItem),
		resetData: make(map[string]resetOtpItem),
	}
}

func (m *otpRepositoryMock) key(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (m *otpRepositoryMock) SetPendingRegistration(ctx context.Context, reg *repository.PendingRegistration, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[m.key(reg.Email)] = pendingRegItem{
		reg:       *reg,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (m *otpRepositoryMock) GetPendingRegistration(ctx context.Context, email string) (*repository.PendingRegistration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[m.key(email)]
	if !exists {
		return nil, nil
	}
	if time.Now().After(item.expiresAt) {
		return nil, nil
	}
	return &item.reg, nil
}

func (m *otpRepositoryMock) DeletePendingRegistration(ctx context.Context, email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, m.key(email))
	return nil
}

func (m *otpRepositoryMock) SetResetOTP(ctx context.Context, email, otp string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.resetData[m.key(email)] = resetOtpItem{
		otp:       otp,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (m *otpRepositoryMock) GetResetOTP(ctx context.Context, email string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.resetData[m.key(email)]
	if !exists {
		return "", nil
	}
	if time.Now().After(item.expiresAt) {
		return "", nil
	}
	return item.otp, nil
}

func (m *otpRepositoryMock) DeleteResetOTP(ctx context.Context, email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.resetData, m.key(email))
	return nil
}
