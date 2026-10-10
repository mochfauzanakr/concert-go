package handler

import (
	"encoding/json"
	"net/http"

	"concert-go/internal/config"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/middleware"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	authUsecase       *usecase.AuthUsecase
	googleOAuthConfig *oauth2.Config
}

// NewAuthHandler constructor
func NewAuthHandler(authUsecase *usecase.AuthUsecase, cfg *config.Config) *AuthHandler {
	googleConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthHandler{
		authUsecase:       authUsecase,
		googleOAuthConfig: googleConfig,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	err := h.authUsecase.Register(c.Request.Context(), req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Verification code sent to your email. Please verify OTP to complete registration.", nil, nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	res, err := h.authUsecase.Login(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Login successful", res, nil)
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req request.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	res, err := h.authUsecase.VerifyOTP(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "OTP verification successful", res, nil)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := uuid.New().String()
	url := h.googleOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		util.RespondBadRequest(c, "Authorization code is required", nil)
		return
	}

	token, err := h.googleOAuthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		util.RespondBadRequest(c, "Failed to exchange authorization code", err.Error())
		return
	}

	client := h.googleOAuthConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		util.RespondInternalError(c)
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		util.RespondInternalError(c)
		return
	}

	if googleUser.Email == "" {
		util.RespondBadRequest(c, "Unable to retrieve email from Google account", nil)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	authRes, err := h.authUsecase.OAuthLogin(c.Request.Context(), googleUser.Email, googleUser.Name, "google", googleUser.ID, ipAddress, userAgent)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Google login successful", authRes, nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	res, err := h.authUsecase.RefreshToken(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Token refreshed successfully", res, nil)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req request.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	if err := h.authUsecase.Logout(c.Request.Context(), req); err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Logged out successfully", nil, nil)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	if err := h.authUsecase.ForgotPassword(c.Request.Context(), req); err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Password reset OTP sent to your email", nil, nil)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	if err := h.authUsecase.ResetPassword(c.Request.Context(), req); err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Password has been reset successfully", nil, nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	res, err := h.authUsecase.GetProfile(c.Request.Context(), session.UserID)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Profile fetched successfully", res, nil)
}
