package entity

import (
	"time"
)

type NoteTag struct {
	NoteID    string    `gorm:"primaryKey;column:note_id;type:varchar(36)"`
	TagID     string    `gorm:"primaryKey;column:tag_id;type:varchar(36)"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	Note *SermonNote `gorm:"foreignKey:NoteID;references:ID"`
	Tag  *Tag        `gorm:"foreignKey:TagID;references:ID"`
}

func (NoteTag) TableName() string {
	return "note_tags"
}
