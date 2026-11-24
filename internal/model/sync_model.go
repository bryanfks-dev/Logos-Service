package model

import "time"

type SyncNotesRequest struct {
	LastSyncedAt *time.Time     `json:"last_synced_at,omitempty"`
	Notes        []SyncNoteItem `json:"notes"`
}

type SyncNoteItem struct {
	ID          string          `json:"id"`
	GatheringID *string         `json:"gathering_id,omitempty"`
	SeriesID    *string         `json:"series_id,omitempty"`
	PreacherID  *string         `json:"preacher_id,omitempty"`
	Title       string          `json:"title"`
	SermonDate  *time.Time      `json:"sermon_date,omitempty"`
	PlainText   string          `json:"plain_text,omitempty"`
	ContentJSON string          `json:"content_json,omitempty"`
	Version     int             `json:"version"`
	Tags        []string        `json:"tags,omitempty"`
	Verses      []SyncNoteVerse `json:"verses,omitempty"`
	IsDeleted   bool            `json:"is_deleted,omitempty"`
	UpdatedAt   *time.Time      `json:"updated_at,omitempty"`
}

type SyncNoteVerse struct {
	ID          string `json:"id,omitempty"`
	Translation string `json:"translation"`
	BookID      int    `json:"book_id"`
	Chapter     int    `json:"chapter"`
	VerseStart  int    `json:"verse_start"`
	VerseEnd    *int   `json:"verse_end,omitempty"`
	DisplayRef  string `json:"display_ref"`
}

type SyncNotesResponse struct {
	SyncedAt   time.Time      `json:"synced_at"`
	AppliedIDs []string       `json:"applied_ids"`
	Notes      []SyncNoteItem `json:"notes"`
}
