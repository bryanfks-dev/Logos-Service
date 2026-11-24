package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type SermonNoteEntity struct {
	ID          string         `gorm:"primaryKey;type:varchar(36);column:id"`
	GatheringID *string        `gorm:"column:gathering_id;type:varchar(36);index"`
	SeriesID    *string        `gorm:"column:series_id;type:varchar(36);index"`
	PreacherID  *string        `gorm:"column:preacher_id;type:varchar(36);index"`
	Title       string         `gorm:"column:title;type:varchar(255);not null"`
	SermonDate  *time.Time     `gorm:"column:sermon_date"`
	FilePath    string         `gorm:"column:file_path;type:text"`
	ContentHash string         `gorm:"column:content_hash;type:varchar(64)"`
	PlainText   string         `gorm:"column:plain_text;type:text"`
	Version     int            `gorm:"column:version;default:1"`
	SyncStatus  string         `gorm:"column:sync_status;type:varchar(50)"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	Gathering  *GatheringEntity    `gorm:"foreignKey:GatheringID;references:ID"`
	Series     *SermonSeriesEntity `gorm:"foreignKey:SeriesID;references:ID"`
	Preacher   *PreacherEntity     `gorm:"foreignKey:PreacherID;references:ID"`
	Tags       []TagEntity         `gorm:"many2many:note_tags;foreignKey:ID;joinForeignKey:NoteID;References:ID;joinReferences:TagID"`
	NoteTags   []NoteTagEntity     `gorm:"foreignKey:NoteID;references:ID"`
	NoteVerses []NoteVerseEntity   `gorm:"foreignKey:NoteID;references:ID"`
	NoteFTS    *NotesFTSEntity     `gorm:"foreignKey:NoteID;references:ID"`
}

func (SermonNoteEntity) TableName() string {
	return "sermon_notes"
}

func (e *SermonNoteEntity) BeforeCreate(db *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}
