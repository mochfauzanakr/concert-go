package response

import "concert-go/internal/domain/model"

// PageResponse is a generic container for paginated list data
type PageResponse[T any] struct {
	Items      []T              `json:"items"`
	Pagination model.Pagination `json:"pagination"`
}
