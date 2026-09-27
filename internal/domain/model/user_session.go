package model

import "github.com/google/uuid"

// UserSession holds authenticated session details attached to context
type UserSession struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
	RoleID *int      `json:"roleId,omitempty"`
}
