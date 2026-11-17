package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"maxi-knows/internal/auth"
)

// PasswordResetGate prevents users from bypassing
// the required password change.
func PasswordResetGate(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path

		// Allow pages needed for login and password changes.
		if path == "/login" ||
			path == "/register" ||
			path == "/change-password" ||
			path == "/api/login" ||
			path == "/api/register" ||
			path == "/api/change-password" ||
			path == "/api/logout" ||
			path == "/favicon.ico" ||
			strings.HasPrefix(path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Check whether the visitor has a session.
		cookie, err := r.Cookie("session")
		if errors.Is(err, http.ErrNoCookie) {
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			http.Error(w, "Session error", http.StatusBadRequest)
			return
		}

		// Read the password change status from SQLite.
		var mustChangePassword bool

		err = db.QueryRow(`
			SELECT u.must_change_password
			FROM sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.token_hash = ?
			  AND s.expires_at > ?
		`, auth.HashSessionToken(cookie.Value), time.Now().Unix()).Scan(&mustChangePassword)

		if errors.Is(err, sql.ErrNoRows) {
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Block normal access until password is changed.
		if mustChangePassword {
			if strings.HasPrefix(path, "/api/") {
				http.Error(w, "Password change required", http.StatusForbidden)
				return
			}

			http.Redirect(w, r, "/change-password", http.StatusSeeOther)
			return
		}

		// Password change already completed.
		next.ServeHTTP(w, r)
	})
}
