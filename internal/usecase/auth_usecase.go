package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"concert-go/internal/config"
	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/domain/payload/response"
	"concert-go/internal/exception"
	"concert-go/internal/repository"
	"concert-go/internal/util"

	"github.com/google/uuid"
)

type AuthUsecase struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	otpRepo     repository.OTPRepository
	emailSender util.EmailSender
	cfg         *config.Config
}

// NewAuthUsecase constructor
func NewAuthUsecase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	otpRepo repository.OTPRepository,
	emailSender util.EmailSender,
	cfg *config.Config,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		otpRepo:     otpRepo,
		emailSender: emailSender,
		cfg:         cfg,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, req request.RegisterRequest) error {
	if err := util.ValidateEmailTLD(req.Email); err != nil {
		return err
	}
	if err := util.ValidatePassword(req.Password); err != nil {
		return err
	}

	existing, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return exception.Conflict("email already registered")
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return err
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}

	pending := &repository.PendingRegistration{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		OTP:          otp,
	}

	if err := u.otpRepo.SetPendingRegistration(ctx, pending, 5*time.Minute); err != nil {
		return err
	}

	if u.emailSender != nil {
		_ = u.emailSender.SendRegistrationOTP(req.Email, req.Name, otp)
	}

	return nil
}

func (u *AuthUsecase) Login(ctx context.Context, req request.LoginRequest, ipAddress, userAgent string) (*response.AuthResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, exception.Unauthorized("invalid email or password")
	}

	if user.PasswordHash == nil || !util.CheckPassword(req.Password, *user.PasswordHash) {
		return nil, exception.Unauthorized("invalid email or password")
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

func (u *AuthUsecase) VerifyOTP(ctx context.Context, req request.VerifyOTPRequest, ipAddress, userAgent string) (*response.AuthResponse, error) {
	if err := util.ValidateEmailTLD(req.Email); err != nil {
		return nil, err
	}

	pending, err := u.otpRepo.GetPendingRegistration(ctx, req.Email)
	if err != nil || pending == nil || pending.OTP != req.OTP {
		return nil, exception.Unauthorized("invalid or expired OTP code")
	}

	// Double check user doesn't already exist
	existing, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, exception.Conflict("email already registered")
	}

	defaultRoleID := 4
	user := &entity.User{
		ID:           uuid.New(),
		Name:         pending.Name,
		Email:        pending.Email,
		PasswordHash: &pending.PasswordHash,
		Provider:     "email",
		RoleID:       &defaultRoleID,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	_ = u.otpRepo.DeletePendingRegistration(ctx, req.Email)

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
		return nil, exception.Unauthorized("invalid or expired token")
	}

	user, err := u.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, exception.NotFound("user not found")
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
		return nil, exception.NotFound("user not found")
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

func generateOTP() (string, error) {
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	num := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	return fmt.Sprintf("%06d", num), nil
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
