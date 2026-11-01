package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func SearchAPIHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		language := r.URL.Query().Get("language")

		if !r.URL.Query().Has("q") {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
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
			if err := rows.Scan(
				&result.URL,
				&result.Title,
				&result.Description,
			); err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			results = append(results, result)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": results,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
