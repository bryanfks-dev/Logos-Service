package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type SermonSeries struct {
	ID          string         `gorm:"primaryKey;type:varchar(36);column:id"`
	Name        string         `gorm:"column:name;type:varchar(255);not null"`
	Description string         `gorm:"column:description;type:text"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	SermonNotes []SermonNote `gorm:"foreignKey:SeriesID;references:ID"`
}

func (SermonSeries) TableName() string {
	return "sermon_series"
}

func (e *SermonSeries) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewV7().String()
	}

	return nil
}
