package storage

import "database/sql"

func InitDB(db *sql.DB) error {
	schema := `
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
    );`

	_, err := db.Exec(schema)
	return err
}
