package handler

import (
	"strings"

	"concert-go/internal/domain/payload/request"
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
	idStr := c.Param("id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		util.RespondBadRequest(c, "Invalid event ID format", err.Error())
		return
	}

	res, err := h.eventUsecase.GetPublicEventDetail(c.Request.Context(), eventID)
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Event detail fetched successfully", res, nil)
}
