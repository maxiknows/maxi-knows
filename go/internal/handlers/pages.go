package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"maxi-knows/internal/auth"
	"maxi-knows/internal/storage"
)

// PageData contains data passed to HTML templates.
type PageData struct {
	Username      string
	Flashes       []string
	Error         string
	FormUsername  string
	FormEmail     string
	Query         string
	SearchResults []SearchResult
}

// SearchResult represents a result shown on the search page.
type SearchResult struct {
	URL         string
	Title       string
	Description string
}

func renderTemplate(w http.ResponseWriter, page string, data PageData) {
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/"+page,
	)
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("Could not render template: %v", err)
		http.Error(w, "Could not render template", http.StatusInternalServerError)
	}
}

func pageHandler(db *sql.DB, page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, page, PageData{
			Username: auth.LoggedInUsername(db, r),
		})
	}
}

func SearchPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		data := PageData{
			Username:      auth.LoggedInUsername(db, r),
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
					description = append(description[:400], []rune(" ...")...)
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

// PageHandler renders a static HTML page with the current user's session.
func PageHandler(db *sql.DB, page string) http.HandlerFunc {
	return pageHandler(db, page)
}
