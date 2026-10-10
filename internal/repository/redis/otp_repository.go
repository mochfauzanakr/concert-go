package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"concert-go/internal/repository"

	goredis "github.com/redis/go-redis/v9"
)

type otpRepositoryImpl struct {
	client *goredis.Client
}

func NewOTPRepository(client *goredis.Client) repository.OTPRepository {
	return &otpRepositoryImpl{client: client}
}

func (r *otpRepositoryImpl) key(email string) string {
	return "pending_reg:email:" + strings.ToLower(strings.TrimSpace(email))
}

func (r *otpRepositoryImpl) resetKey(email string) string {
	return "reset_otp:email:" + strings.ToLower(strings.TrimSpace(email))
}

func (r *otpRepositoryImpl) SetPendingRegistration(ctx context.Context, reg *repository.PendingRegistration, ttl time.Duration) error {
	if r.client == nil {
		return errors.New("redis client not initialized")
	}
	data, err := json.Marshal(reg)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(reg.Email), data, ttl).Err()
}

func (r *otpRepositoryImpl) GetPendingRegistration(ctx context.Context, email string) (*repository.PendingRegistration, error) {
	if r.client == nil {
		return nil, errors.New("redis client not initialized")
	}
	val, err := r.client.Get(ctx, r.key(email)).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var reg repository.PendingRegistration
	if err := json.Unmarshal([]byte(val), &reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *otpRepositoryImpl) DeletePendingRegistration(ctx context.Context, email string) error {
	if r.client == nil {
		return errors.New("redis client not initialized")
	}
	return r.client.Del(ctx, r.key(email)).Err()
}

func (r *otpRepositoryImpl) SetResetOTP(ctx context.Context, email, otp string, ttl time.Duration) error {
	if r.client == nil {
		return errors.New("redis client not initialized")
	}
	return r.client.Set(ctx, r.resetKey(email), otp, ttl).Err()
}

func (r *otpRepositoryImpl) GetResetOTP(ctx context.Context, email string) (string, error) {
	if r.client == nil {
		return "", errors.New("redis client not initialized")
	}
	val, err := r.client.Get(ctx, r.resetKey(email)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", nil
	}
	return val, err
}

func (r *otpRepositoryImpl) DeleteResetOTP(ctx context.Context, email string) error {
	if r.client == nil {
		return errors.New("redis client not initialized")
	}
	return r.client.Del(ctx, r.resetKey(email)).Err()
}
