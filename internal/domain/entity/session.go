package entity

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"column:user_id;type:uuid;not null;index" json:"userId"`
	TokenHash string     `gorm:"column:token_hash;type:varchar(255);not null;uniqueIndex" json:"-"`
	IPAddress string     `gorm:"column:ip_address;type:varchar(45)" json:"ipAddress"`
	UserAgent string     `gorm:"column:user_agent;type:text" json:"userAgent"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null" json:"expiresAt"`
	RevokedAt *time.Time `gorm:"column:revoked_at" json:"revokedAt,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}
