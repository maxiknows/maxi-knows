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

	// Start the web server on port 8080.
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}