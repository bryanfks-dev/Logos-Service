package entity

import (
	"time"
)

type NoteTagEntity struct {
	NoteID    string    `gorm:"primaryKey;column:note_id;type:varchar(36)"`
	TagID     string    `gorm:"primaryKey;column:tag_id;type:varchar(36)"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	Note *SermonNoteEntity `gorm:"foreignKey:NoteID;references:ID"`
	Tag  *TagEntity        `gorm:"foreignKey:TagID;references:ID"`
}

func (NoteTagEntity) TableName() string {
	return "note_tags"
}
