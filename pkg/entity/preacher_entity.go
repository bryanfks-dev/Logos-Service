package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type Preacher struct {
	ID         string         `gorm:"primaryKey;type:varchar(36);column:id"`
	Name       string         `gorm:"column:name;type:varchar(255);not null"`
	ChurchName string         `gorm:"column:church_name;type:varchar(255)"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	SermonNotes []SermonNote `gorm:"foreignKey:PreacherID;references:ID"`
}

func (Preacher) TableName() string {
	return "preachers"
}

func (e *Preacher) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewV7().String()
	}

	return nil
}

type PreacherFilter struct {
	NamePrefix     string
	IncludeDeleted bool
}
