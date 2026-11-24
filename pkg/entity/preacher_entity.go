package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type PreacherEntity struct {
	ID         string         `gorm:"primaryKey;type:varchar(36);column:id"`
	Name       string         `gorm:"column:name;type:varchar(255);not null"`
	ChurchName string         `gorm:"column:church_name;type:varchar(255)"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	SermonNotes []SermonNoteEntity `gorm:"foreignKey:PreacherID;references:ID"`
}

func (PreacherEntity) TableName() string {
	return "preachers"
}

func (e *PreacherEntity) BeforeCreate(db *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}
