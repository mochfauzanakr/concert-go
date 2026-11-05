package response

import (
	"time"

	"github.com/google/uuid"
)

type UserListItemResponse struct {
	ID        uuid.UUID `json:"id"`
	RoleID    *int      `json:"roleId,omitempty"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserDetailResponse struct {
	ID        uuid.UUID `json:"id"`
	RoleID    *int      `json:"roleId,omitempty"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	Provider  string    `json:"provider"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}
