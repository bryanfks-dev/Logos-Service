package entity

type BibleFTSEntity struct {
	VerseID     string `gorm:"primaryKey;column:verse_id;type:varchar(36)"`
	Translation string `gorm:"column:translation;type:varchar(50);not null"`
	DisplayRef  string `gorm:"column:display_ref;type:varchar(100);not null"`
	Text        string `gorm:"column:text;type:text;not null"`

	// Relationships
	Verse *BibleVerseEntity `gorm:"foreignKey:VerseID;references:ID"`
}

func (BibleFTSEntity) TableName() string {
	return "bible_fts"
}
