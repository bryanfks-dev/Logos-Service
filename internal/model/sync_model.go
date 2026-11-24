package model

import (
	"time"

	"github.com/bryanfks-dev/selah/pkg/entity"
)

type SyncBatch struct {
	Notes        []*entity.SermonNote
	SermonNotes  []*entity.SermonNote
	Gatherings   []*entity.Gathering
	Series       []*entity.SermonSeries
	SermonSeries []*entity.SermonSeries
	Preachers    []*entity.Preacher
	Tags         []*entity.Tag
}

type SyncPushResult struct {
	SyncedAt          int64
	MissingContentIDs []string
	Conflicts         []string
}

type SyncPullResult struct {
	ServerTime   int64
	Notes        []*entity.SermonNote
	SermonNotes  []*entity.SermonNote
	Gatherings   []*entity.Gathering
	Series       []*entity.SermonSeries
	SermonSeries []*entity.SermonSeries
	Preachers    []*entity.Preacher
	Tags         []*entity.Tag
}

type SyncPushRequest struct {
	Notes        []SyncNoteItem      `json:"notes,omitempty"`
	SermonNotes  []SyncNoteItem      `json:"sermon_notes,omitempty"`
	Gatherings   []SyncGatheringItem `json:"gatherings,omitempty"`
	Series       []SyncSeriesItem    `json:"series,omitempty"`
	SermonSeries []SyncSeriesItem    `json:"sermon_series,omitempty"`
	Preachers    []SyncPreacherItem  `json:"preachers,omitempty"`
	Tags         []SyncTagItem       `json:"tags,omitempty"`
}

type SyncGatheringItem struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ColorHex  string     `json:"color_hex,omitempty"`
	SortOrder int        `json:"sort_order,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type SyncSeriesItem struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type SyncPreacherItem struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	ChurchName string     `json:"church_name,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type SyncTagItem struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type SyncPushResponse struct {
	SyncedAt          int64    `json:"synced_at"`
	MissingContentIDs []string `json:"missing_content_ids"`
	Conflicts         []string `json:"conflicts"`
}

type SyncPullResponse struct {
	ServerTime time.Time               `json:"server_time"`
	Notes      []*SermonNoteResponse   `json:"notes"`
	Gatherings []*GatheringResponse    `json:"gatherings"`
	Series     []*SermonSeriesResponse `json:"series"`
	Preachers  []*PreacherResponse     `json:"preachers"`
	Tags       []*TagResponse          `json:"tags"`
}

type SermonNoteResponse struct {
	ID          string          `json:"id"`
	GatheringID *string         `json:"gathering_id,omitempty"`
	SeriesID    *string         `json:"series_id,omitempty"`
	PreacherID  *string         `json:"preacher_id,omitempty"`
	Title       string          `json:"title"`
	SermonDate  *time.Time      `json:"sermon_date,omitempty"`
	FilePath    string          `json:"file_path,omitempty"`
	ContentHash string          `json:"content_hash,omitempty"`
	PlainText   string          `json:"plain_text,omitempty"`
	ContentJSON string          `json:"content_json,omitempty"`
	Version     int             `json:"version"`
	SyncStatus  string          `json:"sync_status,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Verses      []SyncNoteVerse `json:"verses,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
	IsDeleted   bool            `json:"is_deleted,omitempty"`
}

type GatheringResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ColorHex  string     `json:"color_hex,omitempty"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	IsDeleted bool       `json:"is_deleted,omitempty"`
}

type SermonSeriesResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	IsDeleted   bool       `json:"is_deleted,omitempty"`
}

type PreacherResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	ChurchName string     `json:"church_name,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	IsDeleted  bool       `json:"is_deleted,omitempty"`
}

type TagResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	UsageCount int       `json:"usage_count,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

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
	ContentHash string          `json:"content_hash,omitempty"`
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
