package request

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type SyncPushChangesRequest struct {
	SermonNotes  []SyncNoteChangesItem      `json:"sermon_notes,omitempty"`
	Gatherings   []SyncGatheringChangesItem `json:"gatherings,omitempty"`
	SermonSeries []SyncSeriesChangesItem    `json:"sermon_series,omitempty"`
	Preachers    []SyncPreacherChangesItem  `json:"preachers,omitempty"`
	Tags         []SyncTagChangesItem       `json:"tags,omitempty"`
}

func (r SyncPushChangesRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.SermonNotes),
		validation.Field(&r.Gatherings),
		validation.Field(&r.SermonSeries),
		validation.Field(&r.Preachers),
		validation.Field(&r.Tags),
	)
}

type SyncGatheringChangesItem struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ColorHex  string     `json:"color_hex,omitempty"`
	SortOrder int        `json:"sort_order,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (i SyncGatheringChangesItem) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ID, validation.Length(0, 36)),
		validation.Field(
			&i.Name,
			validation.Required,
			validation.Length(1, 255),
		),
		validation.Field(
			&i.ColorHex,
			validation.Length(0, 7),
			is.HexColor,
		),
		validation.Field(&i.SortOrder, validation.Min(0)),
	)
}

type SyncSeriesChangesItem struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func (i SyncSeriesChangesItem) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ID, validation.Length(0, 36)),
		validation.Field(
			&i.Name,
			validation.Required,
			validation.Length(1, 255),
		),
	)
}

type SyncPreacherChangesItem struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	ChurchName string     `json:"church_name,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

func (i SyncPreacherChangesItem) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ID, validation.Length(0, 36)),
		validation.Field(
			&i.Name,
			validation.Required,
			validation.Length(1, 255),
		),
		validation.Field(&i.ChurchName, validation.Length(0, 255)),
	)
}

type SyncTagChangesItem struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (i SyncTagChangesItem) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ID, validation.Length(0, 36)),
		validation.Field(
			&i.Name,
			validation.Required,
			validation.Length(1, 100),
		),
	)
}

type SyncNoteChangesItem struct {
	ID          string                     `json:"id"`
	GatheringID *string                    `json:"gathering_id,omitempty"`
	SeriesID    *string                    `json:"series_id,omitempty"`
	PreacherID  *string                    `json:"preacher_id,omitempty"`
	Title       string                     `json:"title"`
	SermonDate  *time.Time                 `json:"sermon_date,omitempty"`
	ContentHash string                     `json:"content_hash,omitempty"`
	PlainText   string                     `json:"plain_text,omitempty"`
	ContentJSON string                     `json:"content_json,omitempty"`
	Version     int                        `json:"version"`
	Tags        []string                   `json:"tags,omitempty"`
	Verses      []SyncNoteVerseChangesItem `json:"verses,omitempty"`
	IsDeleted   bool                       `json:"is_deleted,omitempty"`
	CreateadAt  *time.Time                 `json:"created_at,omitempty"`
	UpdatedAt   *time.Time                 `json:"updated_at,omitempty"`
}

func (i SyncNoteChangesItem) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ID, validation.Length(0, 36)),
		validation.Field(&i.GatheringID, validation.Length(0, 36)),
		validation.Field(&i.SeriesID, validation.Length(0, 36)),
		validation.Field(&i.PreacherID, validation.Length(0, 36)),
		validation.Field(
			&i.Title,
			validation.Required,
			validation.Length(1, 255),
		),
		validation.Field(&i.ContentHash, validation.Length(0, 64)),
		validation.Field(&i.Version, validation.Min(0)),
		validation.Field(
			&i.Tags,
			validation.Each(validation.Required, validation.Length(1, 100)),
		),
		validation.Field(&i.Verses),
	)
}

type SyncNoteVerseChangesItem struct {
	ID          string `json:"id,omitempty"`
	Translation string `json:"translation"`
	BookID      int    `json:"book_id"`
	Chapter     int    `json:"chapter"`
	VerseStart  int    `json:"verse_start"`
	VerseEnd    *int   `json:"verse_end,omitempty"`
	DisplayRef  string `json:"display_ref"`
}

func (v SyncNoteVerseChangesItem) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.ID, validation.Length(0, 36)),
		validation.Field(
			&v.Translation,
			validation.Required,
			validation.Length(1, 50),
		),
		validation.Field(&v.BookID, validation.Required, validation.Min(1)),
		validation.Field(&v.Chapter, validation.Required, validation.Min(1)),
		validation.Field(&v.VerseStart, validation.Required, validation.Min(1)),
		validation.Field(&v.VerseEnd, validation.Min(1)),
		validation.Field(
			&v.DisplayRef,
			validation.Required,
			validation.Length(1, 100),
		),
	)
}
