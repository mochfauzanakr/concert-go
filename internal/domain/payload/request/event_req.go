package request

import "concert-go/internal/domain/model"

type GetPublicEventListRequest struct {
	model.BaseListParams
	City     string `json:"city"`
	Category string `json:"category"`
}
