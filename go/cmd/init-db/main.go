package main

import (
	"database/sql"
	"log"

	"maxi-knows/internal/storage"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "../data/whoknows.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = storage.InitDB(db)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database initialized successfully")
}
