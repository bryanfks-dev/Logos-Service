package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type GatheringEntity struct {
	ID        string    `gorm:"primaryKey;type:varchar(36);column:id"`
	Name      string    `gorm:"column:name;type:varchar(255);not null"`
	ColorHex  string    `gorm:"column:color_hex;type:varchar(7)"`
	SortOrder int       `gorm:"column:sort_order;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	SermonNotes []SermonNoteEntity `gorm:"foreignKey:GatheringID;references:ID"`
}

func (GatheringEntity) TableName() string {
	return "gatherings"
}

func (e *GatheringEntity) BeforeCreate(db *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}
