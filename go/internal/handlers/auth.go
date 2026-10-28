package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"maxi-knows/internal/auth"
	"maxi-knows/internal/storage"
)

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}

		if !r.Form.Has("username") ||
			!r.Form.Has("email") ||
			!r.Form.Has("password") ||
			!r.Form.Has("password2") {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{})
			return
		}

		username := r.Form.Get("username")
		email := r.Form.Get("email")
		password := r.Form.Get("password")
		password2 := r.Form.Get("password2")

		if username == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"statusCode": 422, "message": "Username is required",
			})
			return
		}
		if email == "" || !strings.Contains(email, "@") {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"statusCode": 422, "message": "Valid email is required",
			})
			return
		}
		if password == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"statusCode": 422, "message": "Password is required",
			})
			return
		}
		if password != password2 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"statusCode": 422, "message": "Passwords do not match",
			})
			return
		}

		exists, err := storage.UserExists(db, username, email)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if exists {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"statusCode": 409, "message": "Username or email already exists",
			})
			return
		}

		passwordHash, err := auth.HashPassword(password)
		if err != nil {
			http.Error(w, "Could not hash password", http.StatusInternalServerError)
			return
		}

		err = storage.CreateUser(db, username, email, passwordHash)
		if errors.Is(err, storage.ErrUserExists) {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"statusCode": 409, "message": "Username or email already exists",
			})
			return
		}
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}

		if !r.Form.Has("username") || !r.Form.Has("password") {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{})
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")

		user, err := storage.GetUserByUsername(db, username)
		if errors.Is(err, storage.ErrUserNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"statusCode": 401, "message": "Invalid username",
			})
			return
		}
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		ok, needsUpgrade := auth.CheckPassword(user.PasswordHash, password)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"statusCode": 401, "message": "Invalid password",
			})
			return
		}

		if needsUpgrade {
			newHash, err := auth.HashPassword(password)
			if err != nil {
				http.Error(w, "Could not hash password", http.StatusInternalServerError)
				return
			}
			if err := storage.UpdatePassword(db, user.ID, newHash); err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}

		token, tokenHash, err := auth.NewSessionToken()
		if err != nil {
			http.Error(w, "Could not create session", http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().Add(24 * time.Hour)
		if err := storage.CreateSession(db, tokenHash, user.ID, expiresAt); err != nil {
			http.Error(w, "Could not save session", http.StatusInternalServerError)
			return
		}

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
	}
}

func LogoutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err == nil {
			tokenHash := auth.HashSessionToken(cookie.Value)
			if err := storage.DeleteSession(db, tokenHash); err != nil {
				http.Error(w, "Could not log out", http.StatusInternalServerError)
				return
			}
		}

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
	}
}
