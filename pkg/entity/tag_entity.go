package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type TagEntity struct {
	ID        string    `gorm:"primaryKey;type:varchar(36);column:id"`
	Name      string    `gorm:"column:name;type:varchar(100);not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	SermonNotes []SermonNoteEntity `gorm:"many2many:note_tags;foreignKey:ID;joinForeignKey:TagID;References:ID;joinReferences:NoteID"`
	NoteTags    []NoteTagEntity    `gorm:"foreignKey:TagID;references:ID"`
}

func (TagEntity) TableName() string {
	return "tags"
}

func (e *TagEntity) BeforeCreate(db *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}
