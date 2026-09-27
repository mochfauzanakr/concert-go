package usecase_test

import (
	"context"
	"testing"

	"concert-go/internal/config"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/repository/mock"
	"concert-go/internal/usecase"
)

func setupAuthUsecase() usecase.AuthUsecase {
	userRepo := mock.NewUserRepositoryMock()
	sessionRepo := mock.NewSessionRepositoryMock()
	cfg := &config.Config{
		JWTSecret:           "test-secret-key-123",
		JWTAccessExpMinutes: 15,
		JWTRefreshExpDays:   7,
	}
	return usecase.NewAuthUsecase(userRepo, sessionRepo, cfg)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	uc := setupAuthUsecase()
	ctx := context.Background()

	req := request.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "Password123!",
	}

	res, err := uc.Register(ctx, req, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.User.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, res.User.Email)
	}
	if res.Tokens.AccessToken == "" || res.Tokens.RefreshToken == "" {
		t.Errorf("expected access and refresh tokens to be generated")
	}
}

func TestAuthUsecase_Register_DuplicateEmail(t *testing.T) {
	uc := setupAuthUsecase()
	ctx := context.Background()

	req := request.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "Password123!",
	}

	_, _ = uc.Register(ctx, req, "127.0.0.1", "test-agent")
	_, err := uc.Register(ctx, req, "127.0.0.1", "test-agent")
	if err != usecase.ErrEmailAlreadyExists {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	uc := setupAuthUsecase()
	ctx := context.Background()

	regReq := request.RegisterRequest{
		Name:     "Jane Doe",
		Email:    "jane@example.com",
		Password: "SecretPass123!",
	}
	_, err := uc.Register(ctx, regReq, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("setup registration failed: %v", err)
	}

	loginReq := request.LoginRequest{
		Email:    "jane@example.com",
		Password: "SecretPass123!",
	}
	res, err := uc.Login(ctx, loginReq, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}

	if res.User.Email != loginReq.Email {
		t.Errorf("expected email %s, got %s", loginReq.Email, res.User.Email)
	}
}

func TestAuthUsecase_Login_InvalidPassword(t *testing.T) {
	uc := setupAuthUsecase()
	ctx := context.Background()

	regReq := request.RegisterRequest{
		Name:     "Jane Doe",
		Email:    "jane@example.com",
		Password: "SecretPass123!",
	}
	_, _ = uc.Register(ctx, regReq, "127.0.0.1", "test-agent")

	loginReq := request.LoginRequest{
		Email:    "jane@example.com",
		Password: "wrongPassword",
	}
	_, err := uc.Login(ctx, loginReq, "127.0.0.1", "test-agent")
	if err != usecase.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthUsecase_RefreshToken_Success(t *testing.T) {
	uc := setupAuthUsecase()
	ctx := context.Background()

	regReq := request.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "Password123!",
	}
	authRes, err := uc.Register(ctx, regReq, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	refreshReq := request.RefreshTokenRequest{
		RefreshToken: authRes.Tokens.RefreshToken,
	}

	tokens, err := uc.RefreshToken(ctx, refreshReq, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected refresh token success, got %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Errorf("expected new access and refresh token pair")
	}
}
