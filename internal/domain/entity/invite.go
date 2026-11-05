package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	InviteStatusPending  = "PENDING"
	InviteStatusAccepted = "ACCEPTED"
)

type Invite struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email      string     `gorm:"column:email;type:varchar(100);not null;index" json:"email"`
	RoleID     int        `gorm:"column:role_id;type:int;not null" json:"roleId"`
	InvitedBy  uuid.UUID  `gorm:"column:invited_by;type:uuid;not null" json:"invitedBy"`
	TokenHash  string     `gorm:"column:token_hash;type:varchar(255);unique;not null" json:"-"`
	Status     string     `gorm:"column:status;type:varchar(50);default:'PENDING';not null" json:"status"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null" json:"expiresAt"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	AcceptedAt *time.Time `gorm:"column:accepted_at" json:"acceptedAt,omitempty"`
}
