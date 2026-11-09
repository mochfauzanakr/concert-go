package handler

import (
	"mime/multipart"
	"strings"

	"concert-go/internal/domain/payload/request"
	"concert-go/internal/middleware"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrganizerHandler struct {
	organizerUsecase *usecase.OrganizerUsecase
}

func NewOrganizerHandler(organizerUsecase *usecase.OrganizerUsecase) *OrganizerHandler {
	return &OrganizerHandler{organizerUsecase: organizerUsecase}
}

func (h *OrganizerHandler) GetProfile(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	res, err := h.organizerUsecase.GetProfile(c.Request.Context(), session.UserID)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Organizer profile fetched successfully", res, nil)
}

func (h *OrganizerHandler) GetPublicProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.RespondBadRequest(c, "Invalid organizer ID format", err.Error())
		return
	}

	res, err := h.organizerUsecase.GetPublicProfile(c.Request.Context(), id)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Organizer profile fetched successfully", res, nil)
}

func (h *OrganizerHandler) UpsertProfile(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	var req request.UpsertOrganizerProfileRequest
	var logoHeader *multipart.FileHeader

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req.OrgName = c.PostForm("orgName")
		req.Description = c.PostForm("description")
		req.Phone = c.PostForm("phone")
		req.SocialMedia = c.PostForm("socialMedia")

		if file, err := c.FormFile("logo"); err == nil {
			logoHeader = file
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			util.RespondBadRequest(c, "Invalid input payload", err.Error())
			return
		}
	}

	res, err := h.organizerUsecase.UpsertProfile(c.Request.Context(), session.UserID, req, logoHeader)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Organizer profile updated successfully", res, nil)
}

func (h *OrganizerHandler) GetStaffList(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

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

	req := request.GetOrganizerStaffListRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Status:   status,
	}

	res, err := h.organizerUsecase.GetStaffList(c.Request.Context(), session.UserID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Organizer staff fetched successfully", res.Items, &res.Pagination)
}
