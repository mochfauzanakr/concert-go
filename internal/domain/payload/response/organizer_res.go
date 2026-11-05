package response

import (
	"time"

	"github.com/google/uuid"
)

type OrganizerProfileResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	OrgName     string    `json:"orgName"`
	Email       string    `json:"email"`
	LogoURL     *string   `json:"logoUrl,omitempty"`
	Description *string   `json:"description,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	SocialMedia *string   `json:"socialMedia,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
