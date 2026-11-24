package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NoteVerse struct {
	ID          string    `gorm:"primaryKey;type:varchar(36);column:id"`
	NoteID      string    `gorm:"column:note_id;type:varchar(36);not null;index"`
	Translation string    `gorm:"column:translation;type:varchar(50);not null"`
	BookID      int       `gorm:"column:book_id;not null"`
	Chapter     int       `gorm:"column:chapter;not null"`
	VerseStart  int       `gorm:"column:verse_start;not null"`
	VerseEnd    *int      `gorm:"column:verse_end"`
	DisplayRef  string    `gorm:"column:display_ref;type:varchar(100);not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	Note *SermonNote `gorm:"foreignKey:NoteID;references:ID"`
}

func (NoteVerse) TableName() string {
	return "note_verses"
}

func (e *NoteVerse) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}

	return nil
}

type NoteVerseEntity = NoteVerse
