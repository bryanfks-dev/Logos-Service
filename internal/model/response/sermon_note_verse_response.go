package response

type SermonNoteVerseResponse struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
	BookID      int    `json:"book_id"`
	Chapter     int    `json:"chapter"`
	VerseStart  int    `json:"verse_start"`
	VerseEnd    *int   `json:"verse_end"`
	DisplayRef  string `json:"display_ref"`
}
