package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite"

	"maxi-knows/internal/geo"
	"maxi-knows/internal/handlers"
)

const defaultDBPath = "../data/whoknows.db"

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	// sql.Open never touches the file, and SQLite silently creates an empty
	// DB if the path is wrong, so fail loudly instead.
	if _, err := os.Stat(dbPath); err != nil {
		log.Fatalf("database not found at %s: %v", dbPath, err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Older databases were created without this table, which made login
	// return 500. IF NOT EXISTS makes it a no-op where it already exists.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		token_hash TEXT PRIMARY KEY,
		user_id    INTEGER NOT NULL,
		expires_at INTEGER NOT NULL
	)`); err != nil {
		log.Fatalf("could not create sessions table: %v", err)
	}

	handlers.RegisterRoutes(db, geo.NewService())

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           http.NewCrossOriginProtection().Handler(http.DefaultServeMux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Server running on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
