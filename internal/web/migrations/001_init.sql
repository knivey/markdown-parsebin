CREATE TABLE IF NOT EXISTS pastes (
    slug       TEXT PRIMARY KEY,
    title      TEXT NOT NULL DEFAULT '',
    content    TEXT NOT NULL DEFAULT '',
    rendered   TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    language   TEXT NOT NULL DEFAULT 'markdown'
);
