CREATE TABLE IF NOT EXISTS bible_fts (
    verse_id VARCHAR(36) PRIMARY KEY REFERENCES bible_verses(id) ON DELETE CASCADE,
    translation VARCHAR(50) NOT NULL,
    display_ref VARCHAR(100) NOT NULL,
    text TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bible_fts_search ON bible_fts USING gin(
    to_tsvector('english', text)
);
