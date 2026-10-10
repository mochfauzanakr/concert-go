package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoleID       *int           `gorm:"column:role_id;type:int" json:"roleId"`
	Name         string         `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Email        string         `gorm:"column:email;type:varchar(100);unique;not null" json:"email"`
	PasswordHash *string        `gorm:"column:password_hash;type:varchar(255)" json:"-"`
	Provider     string         `gorm:"column:provider;type:varchar(50);default:'email';not null" json:"provider"`
	ProviderID   *string        `gorm:"column:provider_id;type:varchar(255)" json:"providerId,omitempty"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deletedAt,omitempty"`
}
