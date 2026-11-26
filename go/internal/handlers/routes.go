package handlers

import (
	"database/sql"
	"net/http"

	"maxi-knows/internal/geo"
	"maxi-knows/internal/monitoring"
)

func RegisterRoutes(db *sql.DB, geoSvc *geo.Service) {
	// HTML pages
	http.HandleFunc("GET /{$}", SearchPageHandler(db))
	http.HandleFunc("GET /register", PageHandler(db, "register.html"))
	http.HandleFunc("GET /login", PageHandler(db, "login.html"))
	http.HandleFunc("GET /about", PageHandler(db, "about.html"))

	// Password change page for security breaches
	http.HandleFunc("GET /change-password", PageHandler(db, "change-password.html"))
	http.HandleFunc("POST /api/change-password", ChangePasswordHandler(db))

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

	// Monitoring endpoint
	http.HandleFunc("GET /metrics", monitoring.MetricsHandler(db))

}
