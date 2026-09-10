package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// @title, @version, @description, @host/@BasePath
// RegisterUser godoc
// @Summary      Register a new user
// @Description  Creates a user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration payload"
// @Success      201 {object} RegisterResponse
// @Failure      400 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Router       /api/register [post]

// Package main WhoKnows API
//
// @title           WhoKnows API
// @version         1.0
// @description     Search and user registration API for the WhoKnows project.
// @host            localhost:8080
// @BasePath        /

// SearchResponse is returned by GET /api/search
type SearchResponse struct {
	Data []interface{} `json:"data"`
}

// RegisterResponse is returned on successful registration
type RegisterResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// ErrorResponse is returned when a request fails validation
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func main() {
	http.HandleFunc("GET /{$}", handleHome)
	http.HandleFunc("GET /register", handleRegisterPage)
	http.HandleFunc("GET /login", handleLoginPage)
	http.HandleFunc("GET /api/search", handleSearch)
	http.HandleFunc("POST /api/register", handleRegister)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// handleHome godoc
// @Summary      Home page
// @Description  Returns the WhoKnows landing page
// @Tags         pages
// @Produce      html
// @Success      200 {string} string "WhoKnows"
// @Router       / [get]
func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "WhoKnows")
}

// handleRegisterPage godoc
// @Summary      Registration page
// @Description  Returns the registration form page
// @Tags         pages
// @Produce      html
// @Success      200 {string} string "Register"
// @Router       /register [get]
func handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Register")
}

// handleLoginPage godoc
// @Summary      Login page
// @Description  Returns the login form page
// @Tags         pages
// @Produce      html
// @Success      200 {string} string "Login"
// @Router       /login [get]
func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Login")
}

// handleSearch godoc
// @Summary      Search
// @Description  Searches for matches by query, optionally filtered by language
// @Tags         search
// @Produce      json
// @Param        q         query     string  true   "Search query"
// @Param        language  query     string  false  "Language filter"
// @Success      200 {object} SearchResponse
// @Failure      422 {object} ErrorResponse
// @Router       /api/search [get]
func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	language := r.URL.Query().Get("language")

	//Replace with actual result from db later:
	_ = q
	_ = language

	if !r.URL.Query().Has("q") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)

		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: 422,
			Message:    "q is required",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(SearchResponse{
		//replace with actual matches later:
		Data: []interface{}{},
	})
}

// handleRegister godoc
// @Summary      Register a new user
// @Description  Creates a new user account from form data
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        username  formData  string  true  "Username"
// @Param        email     formData  string  true  "Email address"
// @Param        password  formData  string  true  "Password"
// @Success      200 {object} RegisterResponse
// @Failure      422 {object} ErrorResponse
// @Router       /api/register [post]
func handleRegister(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if !r.Form.Has("username") || !r.Form.Has("email") || !r.Form.Has("password") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)

		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(RegisterResponse{
		StatusCode: 200,
		Message:    "Registered",
	})
}
