package handler

import (
	"errors"

	"concert-go/internal/domain/payload/request"
	"concert-go/internal/middleware"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

// AuthHandler contract
type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	Me(c *gin.Context)
}

type authHandlerImpl struct {
	authUsecase usecase.AuthUsecase
}

// NewAuthHandler constructor
func NewAuthHandler(authUsecase usecase.AuthUsecase) AuthHandler {
	return &authHandlerImpl{
		authUsecase: authUsecase,
	}
}

func (h *authHandlerImpl) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	res, err := h.authUsecase.Register(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			util.RespondBadRequest(c, err.Error(), nil)
			return
		}
		util.RespondInternalError(c)
		return
	}

	util.RespondOK(c, "User registered successfully", res, nil)
}

func (h *authHandlerImpl) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	res, err := h.authUsecase.Login(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			util.RespondUnauthorized(c, err.Error())
			return
		}
		util.RespondInternalError(c)
		return
	}

	util.RespondOK(c, "Login successful", res, nil)
}

func (h *authHandlerImpl) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	res, err := h.authUsecase.RefreshToken(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidToken) || errors.Is(err, usecase.ErrUserNotFound) {
			util.RespondUnauthorized(c, "Invalid or expired session")
			return
		}
		util.RespondInternalError(c)
		return
	}

	util.RespondOK(c, "Token refreshed successfully", res, nil)
}

func (h *authHandlerImpl) Logout(c *gin.Context) {
	var req request.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	if err := h.authUsecase.Logout(c.Request.Context(), req); err != nil {
		util.RespondInternalError(c)
		return
	}

	util.RespondOK(c, "Logged out successfully", nil, nil)
}

func (h *authHandlerImpl) Me(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	res, err := h.authUsecase.GetProfile(c.Request.Context(), session.UserID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			util.RespondUnauthorized(c, "User not found")
			return
		}
		util.RespondInternalError(c)
		return
	}

	util.RespondOK(c, "Profile fetched successfully", res, nil)
}
