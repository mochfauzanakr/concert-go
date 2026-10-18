package repository

import (
	"context"
	"time"
)

type PendingRegistration struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	OTP          string `json:"otp"`
}

type OTPRepository interface {
	SetPendingRegistration(ctx context.Context, reg *PendingRegistration, ttl time.Duration) error
	GetPendingRegistration(ctx context.Context, email string) (*PendingRegistration, error)
	DeletePendingRegistration(ctx context.Context, email string) error

	SetResetOTP(ctx context.Context, email, otp string, ttl time.Duration) error
	GetResetOTP(ctx context.Context, email string) (string, error)
	DeleteResetOTP(ctx context.Context, email string) error
}