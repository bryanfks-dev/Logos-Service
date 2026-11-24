package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type BibleVerse struct {
	ID          string    `gorm:"primaryKey;type:varchar(36);column:id"`
	Translation string    `gorm:"column:translation;type:varchar(50);not null;index:idx_bible_verses_ref"`
	BookID      int       `gorm:"column:book_id;not null;index:idx_bible_verses_ref"`
	BookName    string    `gorm:"column:book_name;type:varchar(100);not null"`
	Chapter     int       `gorm:"column:chapter;not null;index:idx_bible_verses_ref"`
	Verse       int       `gorm:"column:verse;not null;index:idx_bible_verses_ref"`
	Text        string    `gorm:"column:text;type:text;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	BibleFTS *BibleFTS `gorm:"foreignKey:VerseID;references:ID"`
}

func (BibleVerse) TableName() string {
	return "bible_verses"
}

func (e *BibleVerse) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewV7().String()
	}

	return nil
}

type VerseQuery struct {
	Translation string
	BookID      int
	Chapter     int
	VerseStart  int
	VerseEnd    *int
}
