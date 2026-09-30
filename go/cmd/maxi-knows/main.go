package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
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

// cacheEntry holds a cached upstream response and when it expires.
type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// ttlCache is a small in-memory cache with a per-cache TTL.
type ttlCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
}

func newTTLCache(ttl time.Duration) *ttlCache {
	return &ttlCache{items: make(map[string]cacheEntry), ttl: ttl}
}

// get returns the cached value for key if it exists and hasn't expired.
func (c *ttlCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, found := c.items[key]
	c.mu.RUnlock()

	if !found || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

// set stores a value under key until the cache's TTL has passed.
func (c *ttlCache) set(key string, data []byte) {
	c.mu.Lock()
	c.items[key] = cacheEntry{data: data, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

var (
	// Forecasts go stale quickly, place names practically never do.
	weatherCache  = newTTLCache(15 * time.Minute)
	locationCache = newTTLCache(24 * time.Hour)

	// Shared client with a timeout, so a slow upstream can't hang requests.
	httpClient = &http.Client{Timeout: 5 * time.Second}
)

// renderTemplate handles the shared template rendering logic.
// It loads layout.html together with the page-specific template,
// then renders the named "layout" template and sends it to the browser.
func renderTemplate(w http.ResponseWriter, page string, data PageData) {
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/"+page,
	)

	// Stop and return a 500 error if the templates cannot be loaded.
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	// Render the named "layout" template and send the generated HTML
	// to the browser through the ResponseWriter.
	err = tmpl.ExecuteTemplate(w, "layout", data)

	// Return a 500 error if the template could not be rendered.
	if err != nil {

		log.Printf("Could not render template: %v", err)

		http.Error(w, "Could not render template", http.StatusInternalServerError)
	}
}

// pageHandler creates a HTTP handler for pages that only need
// to render a template using the shared layout.
//
// The page parameter decides which page-specific template is rendered.
// For example, pageHandler("about.html") renders about.html together
// with layout.html.
func pageHandler(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Create the data object that is passed to the templates.
		// Username and Flashes are empty for now.
		data := PageData{}

		renderTemplate(w, page, data)
	}
}

// searchPageHandler handles GET requests to the search page.
// It reads the search query from the URL and passes it to the template.
func searchPageHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Query:         r.URL.Query().Get("q"),
		SearchResults: []SearchResult{},
	}

	renderTemplate(w, "search.html", data)
}

// parseCoords reads, validates and rounds the lat/lon query parameters.
// Rounding to 2 decimals (~1 km) keeps the caches effective: forecasts and
// place names don't change at that scale, and visitors share cache entries
// instead of each getting their own.
func parseCoords(r *http.Request) (lat, lon string, ok bool) {
	latF, latErr := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lonF, lonErr := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)

	// Reject missing, non-numeric and out-of-range coordinates.
	if latErr != nil || lonErr != nil ||
		latF < -90 || latF > 90 ||
		lonF < -180 || lonF > 180 {
		return "", "", false
	}

	lat = fmt.Sprintf("%.2f", math.Round(latF*100)/100)
	lon = fmt.Sprintf("%.2f", math.Round(lonF*100)/100)
	return lat, lon, true
}

// fetchBody performs a GET request and returns the body of a 200 response.
func fetchBody(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Nominatim's usage policy requires an identifying User-Agent.
	req.Header.Set("User-Agent", "whoknows-search/1.0 (school project)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// weatherHandler proxies forecast requests to Open-Meteo and caches
// the response per rounded lat/lon, so the homepage widget doesn't
// hit the upstream API on every page load.
func weatherHandler(w http.ResponseWriter, r *http.Request) {
	lat, lon, ok := parseCoords(r)
	if !ok {
		http.Error(w, "valid lat and lon are required", http.StatusBadRequest)
		return
	}

	key := lat + "," + lon

	if data, found := weatherCache.get(key); found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(data)
		return
	}

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current=temperature_2m,weather_code&daily=temperature_2m_max,temperature_2m_min,weather_code&timezone=auto",
		lat, lon,
	)

	body, err := fetchBody(url)
	if err != nil {
		log.Printf("Weather fetch failed: %v", err)
		http.Error(w, "Failed to fetch weather", http.StatusBadGateway)
		return
	}

	weatherCache.set(key, body)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(body)
}

// locationHandler reverse-geocodes lat/lon into a place name using
// OpenStreetMap's Nominatim and returns {"name": "Copenhagen, Denmark"}.
// Results are cached per rounded lat/lon, which also keeps us well inside
// Nominatim's usage policy (max 1 request per second).
func locationHandler(w http.ResponseWriter, r *http.Request) {
	lat, lon, ok := parseCoords(r)
	if !ok {
		http.Error(w, "valid lat and lon are required", http.StatusBadRequest)
		return
	}

	key := lat + "," + lon

	if data, found := locationCache.get(key); found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(data)
		return
	}

	// zoom=10 asks for city-level detail rather than street-level.
	url := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=%s&lon=%s&zoom=10&accept-language=en",
		lat, lon,
	)

	body, err := fetchBody(url)
	if err != nil {
		log.Printf("Location lookup failed: %v", err)
		http.Error(w, "Failed to look up location", http.StatusBadGateway)
		return
	}

	var geo struct {
		Address struct {
			City         string `json:"city"`
			Town         string `json:"town"`
			Village      string `json:"village"`
			Municipality string `json:"municipality"`
			County       string `json:"county"`
			Country      string `json:"country"`
		} `json:"address"`
	}

	if err := json.Unmarshal(body, &geo); err != nil {
		http.Error(w, "Failed to read location response", http.StatusBadGateway)
		return
	}

	// Nominatim uses different keys depending on the size of the place,
	// so take the first one that is filled in.
	name := ""
	for _, part := range []string{
		geo.Address.City, geo.Address.Town, geo.Address.Village,
		geo.Address.Municipality, geo.Address.County,
	} {
		if part != "" {
			name = part
			break
		}
	}

	if geo.Address.Country != "" {
		if name != "" {
			name += ", "
		}
		name += geo.Address.Country
	}

	if name == "" {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	result, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		http.Error(w, "Failed to encode location", http.StatusInternalServerError)
		return
	}

	locationCache.set(key, result)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(result)
}

func main() {
	db, err := sql.Open("sqlite", "../data/whoknows.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// HTML routes

	// GET /SEARCH
	http.HandleFunc("GET /{$}", searchPageHandler)

	// GET /REGISTER
	http.HandleFunc("GET /register", pageHandler("register.html"))

	// GET /LOGIN
	http.HandleFunc("GET /login", pageHandler("login.html"))

	// GET /ABOUT
	// Render the about page using the shared pageHandler.
	http.HandleFunc("GET /about", pageHandler("about.html"))

	// Serve files from the static folder.
	// For example:
	// /static/style.css -> static/style.css
	// /static/monkgroup.png -> static/monkgroup.png
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

		//Check if username is empty
		if username == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "Username is required",
			})
			return
		}

		//Check if email is empty or invalid format
		if email == "" || !strings.Contains(email, "@") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "Valid email is required",
			})
			return
		}

		//check if password is empty
		if password == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "Password is required",
			})
			return
		}

		//check if passwords match
		if password != password2 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "Passwords do not match",
			})
			return
		}

		//check if username or email already exists in db
		var existingUserID int

		err = db.QueryRow(
			"SELECT id FROM users WHERE username = ? OR email = ?",
			username, email,
		).Scan(&existingUserID)

		//If username or email already exists, error message 409 shows
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

		err = row.Scan(&userID, &storedUsername, &storedPassword)

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

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// GET /API/LOGOUT
	http.HandleFunc("GET /api/logout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"statusCode": 200,
			"message":    "Logged out",
		})
	})

	// GET /API/WEATHER
	http.HandleFunc("GET /api/weather", weatherHandler)

	// GET /API/LOCATION
	http.HandleFunc("GET /api/location", locationHandler)

	// Start the web server on port 8080.
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
