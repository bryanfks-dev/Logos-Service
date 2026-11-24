package entity

type BibleFTS struct {
	VerseID     string `gorm:"primaryKey;column:verse_id;type:varchar(36)"`
	Translation string `gorm:"column:translation;type:varchar(50);not null"`
	DisplayRef  string `gorm:"column:display_ref;type:varchar(100);not null"`
	Text        string `gorm:"column:text;type:text;not null"`

	// Relationships
	Verse *BibleVerse `gorm:"foreignKey:VerseID;references:ID"`
}

func (BibleFTS) TableName() string {
	return "bible_fts"
}

type BibleFTSEntity = BibleFTS
type BibleFts = BibleFTS
