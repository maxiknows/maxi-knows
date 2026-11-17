package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"maxi-knows/internal/auth"
	"maxi-knows/internal/storage"
)

// ChangePasswordHandler handles the required password change.
func ChangePasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Find the currently logged-in user.
		username := auth.LoggedInUsername(db, r)

		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := storage.GetUserByUsername(db, username)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check if the user needs to change password.
		if !user.MustChangePassword {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Read the password form.
		r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}

		newPassword := r.PostForm.Get("new_password")
		confirmPassword := r.PostForm.Get("confirm_password")

		// Validate the new password.
		if strings.TrimSpace(newPassword) == "" || len([]byte(newPassword)) > 72 {
			http.Error(w, "Password cannot be empty or only whitespace, or exceed 72 bytes", http.StatusUnprocessableEntity)
			return
		}

		if newPassword != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusUnprocessableEntity)
			return
		}

		// The new password must be different from the old one.
		samePassword, _ := auth.CheckPassword(user.PasswordHash, newPassword)

		if samePassword {
			http.Error(w, "New password must be different from the old password", http.StatusUnprocessableEntity)
			return
		}

		// Hash the new password using the existing bcrypt function.
		newHash, err := auth.HashPassword(newPassword)
		if err != nil {
			http.Error(w, "Could not hash password", http.StatusInternalServerError)
			return
		}

		// Save the new hash and update must_change_password to 0.
		updated, err := storage.CompleteRequiredPasswordChange(db, user.ID, newHash)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if !updated {
			http.Error(w, "Password change is no longer required", http.StatusConflict)
			return
		}

		// Password successfully changed.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
