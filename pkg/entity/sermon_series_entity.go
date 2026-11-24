package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type SermonSeriesEntity struct {
	ID          string         `gorm:"primaryKey;type:varchar(36);column:id"`
	Name        string         `gorm:"column:name;type:varchar(255);not null"`
	Description string         `gorm:"column:description;type:text"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	SermonNotes []SermonNoteEntity `gorm:"foreignKey:SeriesID;references:ID"`
}

func (SermonSeriesEntity) TableName() string {
	return "sermon_series"
}

func (e *SermonSeriesEntity) BeforeCreate(db *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}
