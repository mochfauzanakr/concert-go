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

type EventHandler struct {
	eventUsecase *usecase.EventUsecase
}

func NewEventHandler(eventUsecase *usecase.EventUsecase) *EventHandler {
	return &EventHandler{eventUsecase: eventUsecase}
}

func parseUUIDParam(c *gin.Context, paramName, label string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(paramName))
	if err != nil {
		util.RespondBadRequest(c, "Invalid "+label+" format", err.Error())
		return uuid.Nil, false
	}
	return id, true
}

func parseEventUpsertPayload(c *gin.Context) (request.UpsertEventRequest, *multipart.FileHeader, bool) {
	var req request.UpsertEventRequest
	var bannerHeader *multipart.FileHeader

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req.Name = c.PostForm("name")
		req.Description = c.PostForm("description")
		req.Category = c.PostForm("category")
		req.City = c.PostForm("city")
		req.Location = c.PostForm("location")
		req.StartTime = c.PostForm("startTime")
		req.EndTime = c.PostForm("endTime")
		req.InventoryMode = c.PostForm("inventoryMode")

		if file, err := c.FormFile("banner"); err == nil {
			bannerHeader = file
		}
		return req, bannerHeader, true
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return req, nil, false
	}
	return req, nil, true
}

func (h *EventHandler) GetPublicEvents(c *gin.Context) {
	req := request.GetPublicEventListRequest{
		BaseListParams: util.ParseBaseListParams(c),
		City:           strings.TrimSpace(c.Query("city")),
		Category:       strings.TrimSpace(c.Query("category")),
	}

	res, err := h.eventUsecase.GetPublicEvents(c.Request.Context(), req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Events fetched successfully", res.Items, &res.Pagination)
}

func (h *EventHandler) GetPublicEventDetail(c *gin.Context) {
	eventID, ok := parseUUIDParam(c, "id", "event ID")
	if !ok {
		return
	}

	res, err := h.eventUsecase.GetPublicEventDetail(c.Request.Context(), eventID)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Event detail fetched successfully", res, nil)
}

func (h *EventHandler) GetOrganizerEvents(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	req := request.GetOrganizerEventListRequest{
		BaseListParams: util.ParseBaseListParams(c),
		Status:         strings.TrimSpace(c.Query("status")),
		City:           strings.TrimSpace(c.Query("city")),
		Category:       strings.TrimSpace(c.Query("category")),
	}

	res, err := h.eventUsecase.GetOrganizerEvents(c.Request.Context(), session.UserID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Organizer events fetched successfully", res.Items, &res.Pagination)
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	req, bannerHeader, valid := parseEventUpsertPayload(c)
	if !valid {
		return
	}

	res, err := h.eventUsecase.CreateEvent(c.Request.Context(), session.UserID, req, bannerHeader)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondCreated(c, "Event created successfully", res)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	eventID, ok := parseUUIDParam(c, "id", "event ID")
	if !ok {
		return
	}

	req, bannerHeader, valid := parseEventUpsertPayload(c)
	if !valid {
		return
	}

	res, err := h.eventUsecase.UpdateEvent(c.Request.Context(), session.UserID, eventID, req, bannerHeader)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Event updated successfully", res, nil)
}

func (h *EventHandler) UpdateEventStatus(c *gin.Context) {
	session, ok := middleware.GetUserSession(c)
	if !ok || session == nil {
		util.RespondUnauthorized(c, "Unauthorized")
		return
	}

	eventID, ok := parseUUIDParam(c, "id", "event ID")
	if !ok {
		return
	}

	var req request.UpdateEventStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespondBadRequest(c, "Invalid input payload", err.Error())
		return
	}

	res, err := h.eventUsecase.UpdateEventStatus(c.Request.Context(), session.UserID, eventID, req)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Event status updated successfully", res, nil)
}
