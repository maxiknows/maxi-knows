package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// PageData contains the data that can be passed from the Go backend
// to the HTML templates.
type PageData struct {
	Username string
	Flashes  []string
}

// aboutHandler handles GET requests to /about.
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/about.html",
	)

	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := PageData{}

	err = tmpl.ExecuteTemplate(w, "layout", data)

	if err != nil {
		http.Error(w, "Could not render template", http.StatusInternalServerError)
	}
}

func main() {

	// HTML routes
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "WhoKnows")
	})

	http.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Register")
	})

	http.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Login")
	})

	http.HandleFunc("GET /about", aboutHandler)

	// Static files
	http.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	// API routes
	http.HandleFunc("GET /api/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		language := r.URL.Query().Get("language")

		// Replace with actual result from db later:
		_ = q
		_ = language

		if !r.URL.Query().Has("q") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message":    "q is required",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []interface{}{},
		})
	})

	http.HandleFunc("POST /api/register", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		if !r.Form.Has("username") ||
			!r.Form.Has("email") ||
			!r.Form.Has("password") {

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"statusCode": 200,
			"message":    "Registered",
		})
	})

	http.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		if !r.Form.Has("username") || !r.Form.Has("password") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"statusCode": 200,
			"message":    "Logged in",
		})
	})

	http.HandleFunc("GET /api/logout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"statusCode": 200,
			"message":    "Logged out",
		})
	})

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}