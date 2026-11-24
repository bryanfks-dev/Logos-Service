CREATE TABLE IF NOT EXISTS note_verses (
    id VARCHAR(36) PRIMARY KEY,
    note_id VARCHAR(36) NOT NULL REFERENCES sermon_notes(id) ON DELETE CASCADE,
    translation VARCHAR(50) NOT NULL,
    book_id INT NOT NULL,
    chapter INT NOT NULL,
    verse_start INT NOT NULL,
    verse_end INT,
    display_ref VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_note_verses_note_id ON note_verses(note_id);
