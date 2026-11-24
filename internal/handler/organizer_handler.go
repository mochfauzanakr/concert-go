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

	req := request.GetOrganizerStaffListRequest{
		BaseListParams: util.ParseBaseListParams(c),
		Status:         strings.TrimSpace(c.Query("status")),
	}

	res, err := h.organizerUsecase.GetStaffList(c.Request.Context(), session.UserID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Organizer staff fetched successfully", res.Items, &res.Pagination)
}

func (h *OrganizerHandler) UpdateStaffStatus(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	staffID, err := uuid.Parse(idStr)
	if err != nil {
		util.RespondBadRequest(c, "Invalid staff ID format", err.Error())
		return
	}

	var req request.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	res, err := h.organizerUsecase.UpdateStaffStatus(c.Request.Context(), session.UserID, staffID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Staff status updated successfully", res, nil)
}

func (h *OrganizerHandler) ApplyForPartnership(c *gin.Context) {
	var req request.ApplyOrganizerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	res, err := h.organizerUsecase.ApplyForPartnership(c.Request.Context(), req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondCreated(c, "Organizer application submitted successfully", res)
}

func (h *OrganizerHandler) GetApplications(c *gin.Context) {
	req := request.GetOrganizerApplicationListRequest{
		BaseListParams: util.ParseBaseListParams(c),
		Status:         strings.TrimSpace(c.Query("status")),
	}

	res, err := h.organizerUsecase.GetApplications(c.Request.Context(), req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Organizer applications fetched successfully", res.Items, &res.Pagination)
}

func (h *OrganizerHandler) ReviewApplication(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	appID, err := uuid.Parse(idStr)
	if err != nil {
		util.RespondBadRequest(c, "Invalid application ID format", err.Error())
		return
	}

	var req request.ReviewOrganizerApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	res, err := h.organizerUsecase.ReviewApplication(c.Request.Context(), session.UserID, appID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Organizer application reviewed successfully", res, nil)
}
