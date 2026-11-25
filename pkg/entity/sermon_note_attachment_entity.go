package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type SermonNoteAttachment struct {
	ID             string         `gorm:"primaryKey;type:varchar(36);column:id"`
	NoteID         string         `gorm:"column:note_id;type:varchar(36);not null;index"`
	FileName       string         `gorm:"column:file_name;type:text;not null"`
	MimeType       string         `gorm:"column:mime_type;type:text;not null"`
	FileSize       int64          `gorm:"column:file_size;not null"`
	StoragePath    string         `gorm:"column:storage_path;type:text;not null"`
	ChecksumSHA256 string         `gorm:"column:checksum_sha256;type:text;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime;index"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	Note *SermonNote `gorm:"foreignKey:NoteID;references:ID"`
}

func (SermonNoteAttachment) TableName() string {
	return "note_attachments"
}

func (e *SermonNoteAttachment) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewV7().String()
	}

	return nil
}
