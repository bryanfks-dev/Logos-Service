package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
	"uuid"

	"gorm.io/gorm"
)

type SermonNote struct {
	ID          string         `gorm:"primaryKey;type:varchar(36);column:id"`
	GatheringID *string        `gorm:"column:gathering_id;type:varchar(36);index"`
	SeriesID    *string        `gorm:"column:series_id;type:varchar(36);index"`
	PreacherID  *string        `gorm:"column:preacher_id;type:varchar(36);index"`
	Title       string         `gorm:"column:title;type:varchar(255);not null"`
	SermonDate  time.Time      `gorm:"column:sermon_date"`
	FilePath    string         `gorm:"column:file_path;type:text"`
	ContentHash string         `gorm:"column:content_hash;type:varchar(64)"`
	PlainText   string         `gorm:"column:plain_text;type:text"`
	ContentJSON string         `gorm:"-"`
	Version     int            `gorm:"column:version;default:1"`
	SyncStatus  string         `gorm:"column:sync_status;type:varchar(50)"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	Gathering  *Gathering    `gorm:"foreignKey:GatheringID;references:ID"`
	Series     *SermonSeries `gorm:"foreignKey:SeriesID;references:ID"`
	Preacher   *Preacher     `gorm:"foreignKey:PreacherID;references:ID"`
	Tags       []Tag         `gorm:"many2many:note_tags;foreignKey:ID;joinForeignKey:NoteID;References:ID;joinReferences:TagID"`
	NoteTags   []NoteTag     `gorm:"foreignKey:NoteID;references:ID"`
	Verses     []NoteVerse   `gorm:"foreignKey:NoteID;references:ID"`
	NoteVerses []NoteVerse   `gorm:"foreignKey:NoteID;references:ID"`
	FTS        *NotesFTS     `gorm:"foreignKey:NoteID;references:ID"`
	NoteFTS    *NotesFTS     `gorm:"foreignKey:NoteID;references:ID"`
}

func (SermonNote) TableName() string {
	return "sermon_notes"
}

func (e *SermonNote) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewV7().String()
	}

	if e.FilePath == "" {
		dir := filepath.Join(os.TempDir(), "logos_notes")

		_ = os.MkdirAll(dir, 0o755)

		e.FilePath = filepath.Join(dir, e.ID+".json")
	}

	return nil
}

func (e *SermonNote) HasConflictWith(incoming *SermonNote) bool {
	if e == nil || incoming == nil {
		return false
	}

	return e.Version >= incoming.Version &&
		e.ContentHash != incoming.ContentHash
}

func (e *SermonNote) IsContentMissingFor(incoming *SermonNote) bool {
	if e == nil || incoming == nil {
		return false
	}

	storedHash := e.ContentHash

	if e.ContentJSON != "" {
		sum := sha256.Sum256([]byte(e.ContentJSON))
		storedHash = hex.EncodeToString(sum[:])
	}

	return incoming.ContentHash != storedHash
}

type SermonNoteFilter struct {
	GatheringID    *string
	SeriesID       *string
	PreacherID     *string
	TagID          *string
	StartDate      *time.Time
	EndDate        *time.Time
	IncludeDeleted bool
	UpdatedSince   *time.Time
}

type ScriptureQuery struct {
	Translation string
	BookID      int
	Chapter     int
	VerseStart  int
	VerseEnd    *int
}
