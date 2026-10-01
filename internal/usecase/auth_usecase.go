package usecase

import (
	"context"
	"errors"
	"time"

	"concert-go/internal/config"
	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/domain/payload/response"
	"concert-go/internal/repository"
	"concert-go/internal/util"

	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthUsecase struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	cfg         *config.Config
}

// NewAuthUsecase constructor
func NewAuthUsecase(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, cfg *config.Config) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, req request.RegisterRequest, ipAddress, userAgent string) (*response.AuthResponse, error) {
	if err := util.ValidateEmailTLD(req.Email); err != nil {
		return nil, err
	}
	if err := util.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	existing, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := u.generateAndSaveTokens(ctx, user, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	return &response.AuthResponse{
		User:   toUserResponse(user),
		Tokens: *tokens,
	}, nil
}

func (u *AuthUsecase) Login(ctx context.Context, req request.LoginRequest, ipAddress, userAgent string) (*response.AuthResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !util.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	tokens, err := u.generateAndSaveTokens(ctx, user, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	return &response.AuthResponse{
		User:   toUserResponse(user),
		Tokens: *tokens,
	}, nil
}

func (u *AuthUsecase) RefreshToken(ctx context.Context, req request.RefreshTokenRequest, ipAddress, userAgent string) (*response.TokenResponse, error) {
	tokenHash := util.HashToken(req.RefreshToken)
	session, err := u.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if session == nil || session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	user, err := u.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Revoke old session token
	_ = u.sessionRepo.Revoke(ctx, tokenHash)

	// Issue new token pair
	return u.generateAndSaveTokens(ctx, user, ipAddress, userAgent)
}

func (u *AuthUsecase) Logout(ctx context.Context, req request.LogoutRequest) error {
	tokenHash := util.HashToken(req.RefreshToken)
	return u.sessionRepo.Revoke(ctx, tokenHash)
}

func (u *AuthUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	res := toUserResponse(user)
	return &res, nil
}

func (u *AuthUsecase) generateAndSaveTokens(ctx context.Context, user *entity.User, ipAddress, userAgent string) (*response.TokenResponse, error) {
	now := time.Now()
	accessExpDuration := time.Duration(u.cfg.JWTAccessExpMinutes) * time.Minute
	accessExpiresAt := now.Add(accessExpDuration).Unix()

	claims := util.JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		RoleID:    user.RoleID,
		ExpiresAt: accessExpiresAt,
		IssuedAt:  now.Unix(),
	}

	accessToken, err := util.GenerateJWT(u.cfg.JWTSecret, claims)
	if err != nil {
		return nil, err
	}

	rawRefreshToken := uuid.New().String() + "-" + uuid.New().String()
	tokenHash := util.HashToken(rawRefreshToken)
	refreshExpiresAt := now.Add(time.Duration(u.cfg.JWTRefreshExpDays) * 24 * time.Hour)

	session := &entity.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: refreshExpiresAt,
	}

	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessExpDuration.Seconds()),
	}, nil
}

func toUserResponse(user *entity.User) response.UserResponse {
	return response.UserResponse{
		ID:        user.ID,
		RoleID:    user.RoleID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
