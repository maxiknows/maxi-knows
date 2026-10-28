package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"

	"maxi-knows/internal/geo"
	"maxi-knows/internal/storage"
)

// PageData contains the data that can be passed from the Go backend
// to the HTML templates.
type PageData struct {
	Username      string
	Flashes       []string
	Error         string
	FormUsername  string
	FormEmail     string
	Query         string
	SearchResults []SearchResult
}

// SearchResult represents one result shown on the search page.
type SearchResult struct {
	URL         string
	Title       string
	Description string
}

// renderTemplate handles the shared template rendering logic.
// It loads layout.html together with the page-specific template,
// then renders the named "layout" template and sends it to the browser.
func renderTemplate(w http.ResponseWriter, page string, data PageData) {
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/"+page,
	)
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Could not render template: %v", err)
		http.Error(w, "Could not render template", http.StatusInternalServerError)
	}
}

// pageHandler renders a page with the current user's session data.
func pageHandler(db *sql.DB, page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Username: getLoggedInUsername(db, r),
		}
		renderTemplate(w, page, data)
	}
}

// searchPageHandler handles GET requests to the search page.
// It reads the search query from the URL and passes it to the template.
func searchPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		data := PageData{
			Username:      getLoggedInUsername(db, r),
			Query:         query,
			SearchResults: []SearchResult{},
		}

		if query != "" {
			pages, err := storage.SearchPagesQuery(db, query, "en")
			if err != nil {
				log.Printf("Search failed: %v", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			for _, page := range pages {
				description := []rune(page.Content)

				if len(description) > 400 {
					description = description[:400]
					description = append(description, []rune(" ...")...)
				}

				data.SearchResults = append(data.SearchResults, SearchResult{
					URL:         page.URL,
					Title:       page.Title,
					Description: string(description),
				})
			}
		}

		renderTemplate(w, "search.html", data)
	}
}

// weatherHandler proxies forecast requests to Open-Meteo via the geo
// service, which caches per rounded lat/lon.
func weatherHandler(svc *geo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		lat, lon, ok := geo.ParseCoords(q.Get("lat"), q.Get("lon"))
		if !ok {
			http.Error(w, "valid lat and lon are required", http.StatusBadRequest)
			return
		}

		body, hit, err := svc.Weather(lat, lon)
		if err != nil {
			log.Printf("Weather fetch failed: %v", err)
			http.Error(w, "Failed to fetch weather", http.StatusBadGateway)
			return
		}
		writeCachedJSON(w, body, hit)
	}
}

// locationHandler reverse-geocodes lat/lon into a place name and returns
// {"name": "Copenhagen, Denmark"}.
func locationHandler(svc *geo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		lat, lon, ok := geo.ParseCoords(q.Get("lat"), q.Get("lon"))
		if !ok {
			http.Error(w, "valid lat and lon are required", http.StatusBadRequest)
			return
		}

		body, hit, err := svc.Location(lat, lon)
		if errors.Is(err, geo.ErrLocationNotFound) {
			http.Error(w, "Location not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Location lookup failed: %v", err)
			http.Error(w, "Failed to look up location", http.StatusBadGateway)
			return
		}
		writeCachedJSON(w, body, hit)
	}
}

func writeCachedJSON(w http.ResponseWriter, body []byte, hit bool) {
	w.Header().Set("Content-Type", "application/json")
	if hit {
		w.Header().Set("X-Cache", "HIT")
	} else {
		w.Header().Set("X-Cache", "MISS")
	}
	w.Write(body)
}

func main() {
	db, err := sql.Open("sqlite", "../data/whoknows.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	geoSvc := geo.NewService()

	// HTML routes

	// GET /SEARCH
	http.HandleFunc("GET /{$}", searchPageHandler(db))

	// GET /REGISTER
	http.HandleFunc("GET /register", pageHandler(db, "register.html"))

	// GET /LOGIN
	http.HandleFunc("GET /login", pageHandler(db, "login.html"))

	// GET /ABOUT
	http.HandleFunc("GET /about", pageHandler(db, "about.html"))

	// Serve files from the static folder.
	http.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	// ============================ API routes ============================
	// GET API/SEARCH
	http.HandleFunc("GET /api/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		language := r.URL.Query().Get("language")

		if !r.URL.Query().Has("q") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "q is required",
			})
			return
		}

		if language == "" {
			language = "en"
		}

		rows, err := db.Query(
			"SELECT url, title, content FROM pages WHERE language = ? AND content LIKE ?",
			language, "%"+q+"%",
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		results := []SearchResult{}

		for rows.Next() {
			var result SearchResult

			err = rows.Scan(&result.URL, &result.Title, &result.Description)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			results = append(results, result)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": results,
		})
	})

	// POST /API/REGISTER
	http.HandleFunc("POST /api/register", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		if !r.Form.Has("username") ||
			!r.Form.Has("email") ||
			!r.Form.Has("password") ||
			!r.Form.Has("password2") {

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}

		username := r.Form.Get("username")
		email := r.Form.Get("email")
		password := r.Form.Get("password")
		password2 := r.Form.Get("password2")

		// Check if username is empty
		if username == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "Username is required",
			})
			return
		}

		// Check if email is empty or invalid format
		if email == "" || !strings.Contains(email, "@") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "Valid email is required",
			})
			return
		}

		// Check if password is empty
		if password == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "Password is required",
			})
			return
		}

		// Check if passwords match
		if password != password2 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "Passwords do not match",
			})
			return
		}

		// Check if username or email already exists in db
		var existingUserID int

		err := db.QueryRow(
			"SELECT id FROM users WHERE username = ? OR email = ?",
			username, email,
		).Scan(&existingUserID)

		// If username or email already exists, error message 409 shows
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 409,
				"message":    "Username or email already exists",
			})
			return
		}

		if err != sql.ErrNoRows {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Hash the password using bcrypt
		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			http.Error(w, "Could not hash password", http.StatusInternalServerError)
			return
		}

		// Insert the new user into the database
		_, err = db.Exec(
			"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
			username, email, string(passwordHash),
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// POST /API/LOGIN
	http.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		if !r.Form.Has("username") || !r.Form.Has("password") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")

		row := db.QueryRow(
			"SELECT id, username, password FROM users WHERE username = ?",
			username,
		)

		var userID int
		var storedUsername string
		var storedPassword string

		err := row.Scan(&userID, &storedUsername, &storedPassword)

		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 401,
				"message":    "Invalid username",
			})
			return
		}

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		passwordIsCorrect := false

		// Check if the stored password is already a bcrypt hash
		_, bcryptErr := bcrypt.Cost([]byte(storedPassword))

		if bcryptErr == nil {
			// New user: verify password with bcrypt
			err = bcrypt.CompareHashAndPassword(
				[]byte(storedPassword),
				[]byte(password),
			)

			passwordIsCorrect = err == nil

		} else {
			// Legacy user: verify the old MD5 password
			legacyHash := fmt.Sprintf("%x", md5.Sum([]byte(password)))

			passwordIsCorrect = storedPassword == legacyHash

			// If the old password was correct, upgrade it to bcrypt
			if passwordIsCorrect {
				newHash, err := bcrypt.GenerateFromPassword(
					[]byte(password),
					bcrypt.DefaultCost,
				)
				if err != nil {
					http.Error(w, "Could not hash password", http.StatusInternalServerError)
					return
				}

				_, err = db.Exec(
					"UPDATE users SET password = ? WHERE id = ?",
					string(newHash),
					userID,
				)
				if err != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
				}
			}
		}

		if !passwordIsCorrect {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 401,
				"message":    "Invalid password",
			})
			return
		}

		// Generate a secure session token.
		token, tokenHash, err := generateSessionToken()
		if err != nil {
			http.Error(w, "Could not create session", http.StatusInternalServerError)
			return
		}

		// Store the hashed token in SQLite.
		expiresAt := time.Now().Add(24 * time.Hour)

		if err := storage.CreateSession(db, tokenHash, userID, expiresAt); err != nil {
			http.Error(w, "Could not save session", http.StatusInternalServerError)
			return
		}

		// Send the session token to the browser.
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			Expires:  expiresAt,
			MaxAge:   24 * 60 * 60,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /API/LOGOUT
	http.HandleFunc("POST /api/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")

		if err == nil {
			// Hash the session token to find it in SQLite.
			tokenHash := hashSessionToken(cookie.Value)

			// Remove the session from the database.
			if err := storage.DeleteSession(db, tokenHash); err != nil {
				http.Error(w, "Could not log out", http.StatusInternalServerError)
				return
			}
		}

		// Remove the session cookie from the browser.
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// GET /API/WEATHER
	http.HandleFunc("GET /api/weather", weatherHandler(geoSvc))

	// GET /API/LOCATION
	http.HandleFunc("GET /api/location", locationHandler(geoSvc))

	// Protect the application against cross-origin requests.
	csrf := http.NewCrossOriginProtection()

	// Start the web server on port 8080.
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", csrf.Handler(http.DefaultServeMux)))
}
