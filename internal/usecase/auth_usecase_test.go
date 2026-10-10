package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"concert-go/internal/config"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/exception"
	"concert-go/internal/repository"
	"concert-go/internal/repository/mock"
	"concert-go/internal/usecase"
)

type mockEmailSender struct{}

func (m *mockEmailSender) SendEmail(to []string, subject string, body string, isHTML bool) error {
	return nil
}

func (m *mockEmailSender) SendRegistrationOTP(to, name, otp string) error {
	return nil
}

func (m *mockEmailSender) SendPasswordResetOTP(to, otp string) error {
	return nil
}

func setupAuthUsecase() (*usecase.AuthUsecase, repository.OTPRepository) {
	userRepo := mock.NewUserRepositoryMock()
	sessionRepo := mock.NewSessionRepositoryMock()
	otpRepo := mock.NewOTPRepositoryMock()
	emailSender := &mockEmailSender{}
	cfg := &config.Config{
		JWTSecret:           "test-secret-key-123",
		JWTAccessExpMinutes: 15,
		JWTRefreshExpDays:   7,
	}
	return usecase.NewAuthUsecase(userRepo, sessionRepo, otpRepo, emailSender, cfg), otpRepo
}

func TestAuthUsecase_Register_And_VerifyOTP_Success(t *testing.T) {
	uc, otpRepo := setupAuthUsecase()
	ctx := context.Background()

	req := request.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "Password123!",
	}

	err := uc.Register(ctx, req)
	if err != nil {
		t.Fatalf("expected no error on register, got %v", err)
	}

	pending, err := otpRepo.GetPendingRegistration(ctx, req.Email)
	if err != nil || pending == nil {
		t.Fatalf("expected pending registration in Redis, got %v", err)
	}

	// Verify with invalid OTP
	_, err = uc.VerifyOTP(ctx, request.VerifyOTPRequest{
		Email: req.Email,
		OTP:   "000000",
	}, "127.0.0.1", "test-agent")

	var appErr *exception.AppException
	if !errors.As(err, &appErr) || appErr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for bad OTP, got %v", err)
	}

	// Verify with correct OTP
	res, err := uc.VerifyOTP(ctx, request.VerifyOTPRequest{
		Email: req.Email,
		OTP:   pending.OTP,
	}, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected verify success, got %v", err)
	}

	if res.User.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, res.User.Email)
	}
	if res.User.RoleID == nil || *res.User.RoleID != 4 {
		t.Errorf("expected role ID to be 4 on registration, got %v", res.User.RoleID)
	}
	if res.Tokens.AccessToken == "" || res.Tokens.RefreshToken == "" {
		t.Errorf("expected access and refresh tokens to be generated")
	}

	// Now login with password
	loginRes, err := uc.Login(ctx, request.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if loginRes.User.Email != req.Email {
		t.Errorf("expected login email %s, got %s", req.Email, loginRes.User.Email)
	}
}

func TestAuthUsecase_Register_DuplicateEmail(t *testing.T) {
	uc, otpRepo := setupAuthUsecase()
	ctx := context.Background()

	req := request.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "Password123!",
	}

	_ = uc.Register(ctx, req)
	pending, _ := otpRepo.GetPendingRegistration(ctx, req.Email)
	_, _ = uc.VerifyOTP(ctx, request.VerifyOTPRequest{Email: req.Email, OTP: pending.OTP}, "127.0.0.1", "test-agent")

	err := uc.Register(ctx, req)
	var appErr *exception.AppException
	if !errors.As(err, &appErr) || appErr.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate email, got %v", err)
	}
}

func TestAuthUsecase_Login_InvalidPassword(t *testing.T) {
	uc, otpRepo := setupAuthUsecase()
	ctx := context.Background()

	regReq := request.RegisterRequest{
		Name:     "Jane Doe",
		Email:    "jane@example.com",
		Password: "SecretPass123!",
	}
	_ = uc.Register(ctx, regReq)
	pending, _ := otpRepo.GetPendingRegistration(ctx, regReq.Email)
	_, _ = uc.VerifyOTP(ctx, request.VerifyOTPRequest{Email: regReq.Email, OTP: pending.OTP}, "127.0.0.1", "test-agent")

	loginReq := request.LoginRequest{
		Email:    "jane@example.com",
		Password: "wrongPassword",
	}
	_, err := uc.Login(ctx, loginReq, "127.0.0.1", "test-agent")
	var appErr *exception.AppException
	if !errors.As(err, &appErr) || appErr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for invalid password, got %v", err)
	}
}

func TestAuthUsecase_ForgotPassword_And_ResetPassword_Success(t *testing.T) {
	uc, otpRepo := setupAuthUsecase()
	ctx := context.Background()

	// 1. Register and verify user first
	regReq := request.RegisterRequest{
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "OldPassword123!",
	}
	_ = uc.Register(ctx, regReq)
	pending, _ := otpRepo.GetPendingRegistration(ctx, regReq.Email)
	_, _ = uc.VerifyOTP(ctx, request.VerifyOTPRequest{Email: regReq.Email, OTP: pending.OTP}, "127.0.0.1", "test-agent")

	// 2. Forgot password request
	err := uc.ForgotPassword(ctx, request.ForgotPasswordRequest{Email: "bob@example.com"})
	if err != nil {
		t.Fatalf("expected forgot password success, got %v", err)
	}

	otp, err := otpRepo.GetResetOTP(ctx, "bob@example.com")
	if err != nil || otp == "" {
		t.Fatalf("expected reset OTP stored, got %v", err)
	}

	// 3. Reset password with bad OTP
	err = uc.ResetPassword(ctx, request.ResetPasswordRequest{
		Email:       "bob@example.com",
		OTP:         "000000",
		NewPassword: "NewPassword123!",
	})
	var appErr *exception.AppException
	if !errors.As(err, &appErr) || appErr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for bad reset OTP, got %v", err)
	}

	// 4. Reset password with correct OTP
	err = uc.ResetPassword(ctx, request.ResetPasswordRequest{
		Email:       "bob@example.com",
		OTP:         otp,
		NewPassword: "NewPassword123!",
	})
	if err != nil {
		t.Fatalf("expected reset password success, got %v", err)
	}

	// 5. Try login with old password (should fail)
	_, err = uc.Login(ctx, request.LoginRequest{
		Email:    "bob@example.com",
		Password: "OldPassword123!",
	}, "127.0.0.1", "test-agent")
	if !errors.As(err, &appErr) || appErr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for old password, got %v", err)
	}

	// 6. Try login with new password (should succeed)
	loginRes, err := uc.Login(ctx, request.LoginRequest{
		Email:    "bob@example.com",
		Password: "NewPassword123!",
	}, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected login with new password success, got %v", err)
	}
	if loginRes.User.Email != "bob@example.com" {
		t.Errorf("expected email bob@example.com, got %s", loginRes.User.Email)
	}
}

func TestAuthUsecase_RefreshToken_Success(t *testing.T) {
	uc, otpRepo := setupAuthUsecase()
	ctx := context.Background()

	regReq := request.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "Password123!",
	}
	_ = uc.Register(ctx, regReq)
	pending, _ := otpRepo.GetPendingRegistration(ctx, regReq.Email)
	authRes, err := uc.VerifyOTP(ctx, request.VerifyOTPRequest{Email: regReq.Email, OTP: pending.OTP}, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("register and verify failed: %v", err)
	}

	refreshReq := request.RefreshTokenRequest{
		RefreshToken: authRes.Tokens.RefreshToken,
	}

	time.Sleep(10 * time.Millisecond)
	tokens, err := uc.RefreshToken(ctx, refreshReq, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected refresh token success, got %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Errorf("expected new access and refresh token pair")
	}
}
