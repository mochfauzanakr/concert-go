package entity

import (
	"time"

	"gorm.io/gorm"
)

// BaseEntity provides common fields for all GORM entities.
// DB columns use snake_case per rules.
type BaseEntity struct {
	ID        uint           `gorm:"primaryKey;column:id" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deletedAt,omitempty"`
}
