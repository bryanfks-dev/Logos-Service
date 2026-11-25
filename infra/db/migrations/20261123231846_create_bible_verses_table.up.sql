CREATE TABLE IF NOT EXISTS bible_verses (
    id VARCHAR(36) PRIMARY KEY,
    translation VARCHAR(50) NOT NULL,
    book_id INT NOT NULL,
    book_name VARCHAR(100) NOT NULL,
    chapter INT NOT NULL,
    verse INT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bible_verses_ref ON bible_verses(translation, book_id, chapter, verse);
