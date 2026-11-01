CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS pages (
    title TEXT PRIMARY KEY UNIQUE,
    url TEXT NOT NULL UNIQUE,
    language TEXT NOT NULL CHECK(language IN ('en', 'da')) DEFAULT 'en',
    last_updated TIMESTAMP,
    content TEXT NOT NULL
);

-- One row, so the search test goes through a real result instead of an empty list.
INSERT OR IGNORE INTO pages (title, url, language, content)
VALUES ('Smoke Test Page', 'https://example.com/smoke', 'en', 'A page that exists so the test search has a hit.');