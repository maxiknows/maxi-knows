package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "modernc.org/sqlite"

	"maxi-knows/internal/geo"
	"maxi-knows/internal/handlers"
)

func main() {
	db, err := sql.Open("sqlite", "../data/whoknows.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	handlers.RegisterRoutes(db, geo.NewService())

	csrf := http.NewCrossOriginProtection()

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", csrf.Handler(http.DefaultServeMux)))
}
