package response

import "concert-go/internal/domain/model"

// Pagination alias to domain model.Pagination
type Pagination = model.Pagination

// GlobalResponse is the standard format required by rules
type GlobalResponse struct {
	Message    string            `json:"message"`
	Data       interface{}       `json:"data,omitempty"`
	Pagination *model.Pagination `json:"pagination,omitempty"`
	ReqID      string            `json:"reqId"`
	Status     string            `json:"status"` // "T" or "F"
}

// ErrorResponse is for 4xx and 5xx errors
type ErrorResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
