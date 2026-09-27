package entity

import (
	"time"
)

type Role struct {
	ID        int       `gorm:"primary_key;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(50);unique;not null" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}