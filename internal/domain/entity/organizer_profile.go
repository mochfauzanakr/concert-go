package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrganizerProfile struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"column:user_id;type:uuid;uniqueIndex;not null" json:"userId"`
	OrgName     string    `gorm:"column:org_name;type:varchar(100);not null" json:"orgName"`
	LogoURL     *string   `gorm:"column:logo_url;type:varchar(255)" json:"logoUrl,omitempty"`
	Description *string   `gorm:"column:description;type:text" json:"description,omitempty"`
	Phone       *string   `gorm:"column:phone;type:varchar(20)" json:"phone,omitempty"`
	SocialMedia *string   `gorm:"column:social_media;type:varchar(255)" json:"socialMedia,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}
