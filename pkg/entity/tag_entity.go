package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type Tag struct {
	ID         string    `gorm:"primaryKey;type:varchar(36);column:id"`
	Name       string    `gorm:"column:name;type:varchar(100);not null;uniqueIndex"`
	UsageCount int       `gorm:"-"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	SermonNotes []SermonNote `gorm:"many2many:note_tags;foreignKey:ID;joinForeignKey:TagID;References:ID;joinReferences:NoteID"`
	NoteTags    []NoteTag    `gorm:"foreignKey:TagID;references:ID"`
}

func (Tag) TableName() string {
	return "tags"
}

func (e *Tag) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewV7().String()
	}

	return nil
}
