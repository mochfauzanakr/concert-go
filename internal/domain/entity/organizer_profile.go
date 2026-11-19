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

type OrganizerStaff struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizerID uuid.UUID `gorm:"column:organizer_id;type:uuid;uniqueIndex:idx_organizer_staff;not null" json:"organizerId"`
	UserID      uuid.UUID `gorm:"column:user_id;type:uuid;uniqueIndex:idx_organizer_staff;not null" json:"userId"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

const (
	ApplicationStatusPending  = "PENDING"
	ApplicationStatusApproved = "APPROVED"
	ApplicationStatusRejected = "REJECTED"
)

type OrganizerApplication struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PicName         string     `gorm:"column:pic_name;type:varchar(100);not null" json:"picName"`
	Email           string     `gorm:"column:email;type:varchar(150);index;not null" json:"email"`
	Phone           string     `gorm:"column:phone;type:varchar(20);not null" json:"phone"`
	OrgName         string     `gorm:"column:org_name;type:varchar(100);not null" json:"orgName"`
	SocialMedia     *string    `gorm:"column:social_media;type:varchar(255)" json:"socialMedia,omitempty"`
	Description     string     `gorm:"column:description;type:text;not null" json:"description"`
	Status          string     `gorm:"column:status;type:varchar(20);default:'PENDING';not null" json:"status"`
	RejectionReason *string    `gorm:"column:rejection_reason;type:text" json:"rejectionReason,omitempty"`
	ReviewedBy      *uuid.UUID `gorm:"column:reviewed_by;type:uuid" json:"reviewedBy,omitempty"`
	ReviewedAt      *time.Time `gorm:"column:reviewed_at" json:"reviewedAt,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}
