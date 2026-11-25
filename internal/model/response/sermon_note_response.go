package response

import (
	"time"
)

type SermonNoteResponse struct {
	ID          string                    `json:"id"`
	GatheringID *string                   `json:"gathering_id,omitempty"`
	SeriesID    *string                   `json:"series_id,omitempty"`
	PreacherID  *string                   `json:"preacher_id,omitempty"`
	Title       string                    `json:"title"`
	SermonDate  *time.Time                `json:"sermon_date,omitempty"`
	FilePath    string                    `json:"file_path,omitempty"`
	ContentHash string                    `json:"content_hash,omitempty"`
	PlainText   string                    `json:"plain_text,omitempty"`
	ContentJSON string                    `json:"content_json,omitempty"`
	Version     int                       `json:"version"`
	SyncStatus  string                    `json:"sync_status,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Verses      []SermonNoteVerseResponse `json:"verses,omitempty"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	DeletedAt   *time.Time                `json:"deleted_at,omitempty"`
	IsDeleted   bool                      `json:"is_deleted,omitempty"`
}
