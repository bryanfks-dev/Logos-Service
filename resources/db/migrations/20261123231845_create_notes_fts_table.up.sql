CREATE TABLE IF NOT EXISTS notes_fts (
    note_id VARCHAR(36) PRIMARY KEY REFERENCES sermon_notes(id) ON DELETE CASCADE,
    title TEXT,
    preacher_name VARCHAR(255),
    plain_text TEXT
);

CREATE INDEX IF NOT EXISTS idx_notes_fts_search ON notes_fts USING gin(
    to_tsvector('english', coalesce(title, '') || ' ' || coalesce(preacher_name, '') || ' ' || coalesce(plain_text, ''))
);
