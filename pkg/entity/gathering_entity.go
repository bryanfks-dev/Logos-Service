package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type Gathering struct {
	ID        string         `gorm:"primaryKey;type:varchar(36);column:id"`
	Name      string         `gorm:"column:name;type:varchar(255);not null"`
	ColorHex  string         `gorm:"column:color_hex;type:varchar(7)"`
	SortOrder int            `gorm:"column:sort_order;default:0"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	SermonNotes []SermonNote `gorm:"foreignKey:GatheringID;references:ID"`
}

func (Gathering) TableName() string {
	return "gatherings"
}

func (e *Gathering) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewV7().String()
	}

	return nil
}
