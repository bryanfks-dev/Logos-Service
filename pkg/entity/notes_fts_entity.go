package entity

type NotesFTSEntity struct {
	NoteID       string `gorm:"primaryKey;column:note_id;type:varchar(36)"`
	Title        string `gorm:"column:title;type:text"`
	PreacherName string `gorm:"column:preacher_name;type:varchar(255)"`
	PlainText    string `gorm:"column:plain_text;type:text"`

	// Relationships
	Note *SermonNoteEntity `gorm:"foreignKey:NoteID;references:ID"`
}

func (NotesFTSEntity) TableName() string {
	return "notes_fts"
}
