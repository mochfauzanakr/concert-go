package response

import (
	"time"

	"github.com/google/uuid"
)

type InviteResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expiresAt"`
}
