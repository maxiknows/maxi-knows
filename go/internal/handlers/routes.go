package handlers

import (
	"database/sql"
	"net/http"

	"maxi-knows/internal/geo"
)

func RegisterRoutes(db *sql.DB, geoSvc *geo.Service) {
	// HTML pages
	http.HandleFunc("GET /{$}", SearchPageHandler(db))
	http.HandleFunc("GET /register", PageHandler(db, "register.html"))
	http.HandleFunc("GET /login", PageHandler(db, "login.html"))
	http.HandleFunc("GET /about", PageHandler(db, "about.html"))

	// Static files
	http.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	// Search API
	http.HandleFunc("GET /api/search", SearchAPIHandler(db))

	// Authentication API
	http.HandleFunc("POST /api/register", RegisterHandler(db))
	http.HandleFunc("POST /api/login", LoginHandler(db))
	http.HandleFunc("POST /api/logout", LogoutHandler(db))

	// Geolocation API
	http.HandleFunc("GET /api/weather", WeatherHandler(geoSvc))
	http.HandleFunc("GET /api/location", LocationHandler(geoSvc))
}
