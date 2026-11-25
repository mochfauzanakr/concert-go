package response

import (
	"time"

	"github.com/google/uuid"
)

type EventListItemResponse struct {
	ID            uuid.UUID `json:"id"`
	OrganizerID   uuid.UUID `json:"organizerId"`
	Name          string    `json:"name"`
	BannerURL     *string   `json:"bannerUrl,omitempty"`
	Category      string    `json:"category"`
	City          string    `json:"city"`
	Location      string    `json:"location"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Status        string    `json:"status"`
	InventoryMode string    `json:"inventoryMode"`
	CreatedAt     time.Time `json:"createdAt"`
}

type EventOrganizerSummaryResponse struct {
	UserID    uuid.UUID  `json:"userId"`
	ProfileID *uuid.UUID `json:"profileId,omitempty"`
	OrgName   string     `json:"orgName"`
	LogoURL   *string    `json:"logoUrl,omitempty"`
}

type EventCoOrganizerResponse struct {
	UserID     uuid.UUID `json:"userId"`
	Name       string    `json:"name"`
	OrgName    string    `json:"orgName"`
	LogoURL    *string   `json:"logoUrl,omitempty"`
	AssignedAt time.Time `json:"assignedAt"`
}

type EventArtistResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Bio       *string   `json:"bio,omitempty"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
}

type EventTicketSummaryResponse struct {
	TotalAvailable int      `json:"totalAvailable"`
	LowestPrice    *float64 `json:"lowestPrice"`
	SoldOut        bool     `json:"soldOut"`
}

type EventDetailResponse struct {
	ID            uuid.UUID                      `json:"id"`
	Name          string                         `json:"name"`
	Description   string                         `json:"description"`
	BannerURL     *string                        `json:"bannerUrl,omitempty"`
	Category      string                         `json:"category"`
	City          string                         `json:"city"`
	Location      string                         `json:"location"`
	StartTime     time.Time                      `json:"startTime"`
	EndTime       time.Time                      `json:"endTime"`
	Status        string                         `json:"status"`
	InventoryMode string                         `json:"inventoryMode"`
	LeadOrganizer *EventOrganizerSummaryResponse `json:"leadOrganizer,omitempty"`
	CoOrganizers  []EventCoOrganizerResponse     `json:"coOrganizers"`
	Artists       []EventArtistResponse          `json:"artists"`
	TicketSummary EventTicketSummaryResponse     `json:"ticketSummary"`
	CreatedAt     time.Time                      `json:"createdAt"`
	UpdatedAt     time.Time                      `json:"updatedAt"`
}

type EventTeamMemberResponse struct {
	ID          uuid.UUID `json:"id"`
	EventID     uuid.UUID `json:"eventId"`
	UserID      uuid.UUID `json:"userId"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	OrgName     *string   `json:"orgName,omitempty"`
	AvatarURL   *string   `json:"avatarUrl,omitempty"`
	RoleInEvent string    `json:"roleInEvent"`
	Status      string    `json:"status"`
	AssignedAt  time.Time `json:"assignedAt"`
}
