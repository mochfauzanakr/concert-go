package handler

import (
	"net/http"

	"concert-go/internal/domain/payload/request"
	"concert-go/internal/middleware"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(userUsecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// InviteUser handles user invitation dispatch (SUPER_ADMIN, ADMIN, ORGANIZER)
func (h *UserHandler) InviteUser(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	var req request.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	res, err := h.userUsecase.InviteUser(c.Request.Context(), session.UserID, session.RoleID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Invitation sent successfully",
		"data":    res,
		"status":  "T",
	})
}

// AcceptInvite activates invited user account
func (h *UserHandler) AcceptInvite(c *gin.Context) {
	var req request.AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	if req.Token == "" {
		req.Token = c.Query("token")
	}
	if req.Token == "" {
		util.RespondBadRequest(c, "Invitation token is required", nil)
		return
	}

	user, err := h.userUsecase.AcceptInvite(c.Request.Context(), req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account activated successfully",
		"data":    user,
		"status":  "T",
	})
}

// ValidateInvite checks if an invite token is valid and returns its metadata
func (h *UserHandler) ValidateInvite(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		util.RespondBadRequest(c, "Token query parameter is required", nil)
		return
	}

	res, err := h.userUsecase.ValidateInvite(c.Request.Context(), token)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Invitation token is valid", res, nil)
}
