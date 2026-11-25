package entity

import (
	"time"

	"github.com/google/uuid"
)

type Artist struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(150);not null;index" json:"name"`
	Bio       *string   `gorm:"column:bio;type:text" json:"bio,omitempty"`
	AvatarURL *string   `gorm:"column:avatar_url;type:varchar(255)" json:"avatarUrl,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

type EventArtist struct {
	EventID  uuid.UUID `gorm:"column:event_id;type:uuid;primaryKey" json:"eventId"`
	ArtistID uuid.UUID `gorm:"column:artist_id;type:uuid;primaryKey" json:"artistId"`
}
