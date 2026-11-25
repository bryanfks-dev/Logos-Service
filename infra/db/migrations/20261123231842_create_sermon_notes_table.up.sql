CREATE TABLE IF NOT EXISTS sermon_notes (
    id VARCHAR(36) PRIMARY KEY,
    gathering_id VARCHAR(36) REFERENCES gatherings(id) ON DELETE SET NULL,
    series_id VARCHAR(36) REFERENCES sermon_series(id) ON DELETE SET NULL,
    preacher_id VARCHAR(36) REFERENCES preachers(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    sermon_date TIMESTAMPTZ NOT NULL,
    file_path TEXT,
    content_hash VARCHAR(64),
    plain_text TEXT,
    version INT NOT NULL DEFAULT 1,
    sync_status VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sermon_notes_gathering_id ON sermon_notes(gathering_id);
CREATE INDEX IF NOT EXISTS idx_sermon_notes_series_id ON sermon_notes(series_id);
CREATE INDEX IF NOT EXISTS idx_sermon_notes_preacher_id ON sermon_notes(preacher_id);
CREATE INDEX IF NOT EXISTS idx_sermon_notes_deleted_at ON sermon_notes(deleted_at);
