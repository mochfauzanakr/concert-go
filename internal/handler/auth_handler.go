package handler

import (
	"concert-go/internal/config"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/middleware"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

// NewAuthHandler constructor
func NewAuthHandler(authUsecase *usecase.AuthUsecase, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
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

func (h *AuthHandler) Me(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok {
		util.RespondUnauthorized(c, "Session not found")
		return
	}

	res, err := h.authUsecase.GetProfile(c.Request.Context(), session.UserID)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "User profile retrieved", res, nil)
}
