package handler

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"concert-go/internal/domain/payload/request"
	"concert-go/internal/middleware"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(userUsecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

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

func (h *UserHandler) GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.RespondBadRequest(c, "Invalid user ID format", err.Error())
		return
	}

	user, err := h.userUsecase.GetUserDetail(c.Request.Context(), id)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "User fetched successfully", user, nil)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	page := util.ParseQueryInt(c, "page", 1)
	pageSize := util.ParseQueryInt(c, "pageSize", 10)
	if c.Query("pageSize") == "" && c.Query("limit") != "" {
		pageSize = util.ParseQueryInt(c, "limit", 10)
	}

	search := strings.TrimSpace(c.Query("search"))
	if search == "" {
		search = strings.TrimSpace(c.Query("searchQuery"))
	}
	status := strings.TrimSpace(c.Query("status"))

	var roleID *int
	if rawRole := strings.TrimSpace(c.Query("roleId")); rawRole != "" {
		if id, err := strconv.Atoi(rawRole); err == nil {
			roleID = &id
		}
	}

	req := request.GetUserListRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		RoleID:   roleID,
		Status:   status,
	}

	res, err := h.userUsecase.GetAllUsers(c.Request.Context(), req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Users fetched successfully", res.Items, &res.Pagination)
}

func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		util.RespondBadRequest(c, "Invalid user ID format", err.Error())
		return
	}

	var req request.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	user, err := h.userUsecase.UpdateUserStatus(c.Request.Context(), session.UserID, session.RoleID, targetID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "User status updated successfully", user, nil)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	name := c.PostForm("name")
	var avatarHeader *multipart.FileHeader

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, err := c.FormFile("avatar")
		if err == nil {
			avatarHeader = file
		}
	} else {
		var req request.UpdateProfileRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.Name != "" {
			name = req.Name
		}
	}

	user, err := h.userUsecase.UpdateProfile(c.Request.Context(), session.UserID, name, avatarHeader)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Profile updated successfully", user, nil)
}

func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		util.RespondBadRequest(c, "Invalid user ID format", err.Error())
		return
	}

	var req request.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	user, err := h.userUsecase.AdminUpdateUser(c.Request.Context(), session.UserID, session.RoleID, targetID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "User updated successfully", user, nil)
}
