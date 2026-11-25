package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	EventStatusDraft     = "DRAFT"
	EventStatusPublished = "PUBLISHED"
	EventStatusCompleted = "COMPLETED"
	EventStatusCancelled = "CANCELLED"

	InventoryModeOnlineOnly  = "ONLINE_ONLY"
	InventoryModeDualChannel = "DUAL_CHANNEL"

	EventRoleCoOrganizer = "CO_ORGANIZER"
	EventRoleStaff       = "STAFF"
)

type Event struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizerID   uuid.UUID      `gorm:"column:organizer_id;type:uuid;not null;index" json:"organizerId"`
	Name          string         `gorm:"column:name;type:varchar(150);not null" json:"name"`
	Description   string         `gorm:"column:description;type:text;not null" json:"description"`
	BannerURL     *string        `gorm:"column:banner_url;type:varchar(255)" json:"bannerUrl,omitempty"`
	Category      string         `gorm:"column:category;type:varchar(100);not null;index" json:"category"`
	City          string         `gorm:"column:city;type:varchar(100);not null;index" json:"city"`
	Location      string         `gorm:"column:location;type:varchar(255);not null" json:"location"`
	StartTime     time.Time      `gorm:"column:start_time;not null;index" json:"startTime"`
	EndTime       time.Time      `gorm:"column:end_time;not null" json:"endTime"`
	Status        string         `gorm:"column:status;type:varchar(50);default:'DRAFT';not null;index" json:"status"`
	InventoryMode string         `gorm:"column:inventory_mode;type:varchar(50);default:'ONLINE_ONLY';not null" json:"inventoryMode"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deletedAt,omitempty"`
}

type EventOrganizer struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventID     uuid.UUID `gorm:"column:event_id;type:uuid;uniqueIndex:idx_event_organizer_user;not null" json:"eventId"`
	UserID      uuid.UUID `gorm:"column:user_id;type:uuid;uniqueIndex:idx_event_organizer_user;not null" json:"userId"`
	RoleInEvent string    `gorm:"column:role_in_event;type:varchar(50);not null" json:"roleInEvent"`
	AssignedAt  time.Time `gorm:"column:assigned_at;autoCreateTime" json:"assignedAt"`
}
