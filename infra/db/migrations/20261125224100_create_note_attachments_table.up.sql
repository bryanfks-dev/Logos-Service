CREATE TABLE IF NOT EXISTS note_attachments (
    id VARCHAR(36) PRIMARY KEY,
    note_id VARCHAR(36) NOT NULL REFERENCES sermon_notes(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    checksum_sha256 TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'not_uploaded',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_note_attachments_note_id ON note_attachments(note_id);
CREATE INDEX IF NOT EXISTS idx_note_attachments_updated_at ON note_attachments(updated_at);
