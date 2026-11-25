package request

import "concert-go/internal/domain/model"

type GetPublicEventListRequest struct {
	model.BaseListParams
	City     string `json:"city"`
	Category string `json:"category"`
}

type GetOrganizerEventListRequest struct {
	model.BaseListParams
	Status   string `json:"status"`
	City     string `json:"city"`
	Category string `json:"category"`
}

type UpsertEventRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	City          string `json:"city"`
	Location      string `json:"location"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	InventoryMode string `json:"inventoryMode"`
}

type UpdateEventStatusRequest struct {
	Status string `json:"status"`
}
