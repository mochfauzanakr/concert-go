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

type OrganizerApplicationResponse struct {
	ID              uuid.UUID  `json:"id"`
	PicName         string     `json:"picName"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	OrgName         string     `json:"orgName"`
	SocialMedia     *string    `json:"socialMedia,omitempty"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejectionReason,omitempty"`
	ReviewedBy      *uuid.UUID `json:"reviewedBy,omitempty"`
	ReviewedAt      *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}
