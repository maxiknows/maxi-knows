package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"

	"maxi-knows/internal/storage"
)

// NewSessionToken creates a random token for the cookie and its
// SHA-256 hash for database storage.
func NewSessionToken() (token, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = hex.EncodeToString(b)
	return token, HashSessionToken(token), nil
}

// HashSessionToken returns the SHA-256 hash of a session token.
func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// LoggedInUsername returns the username for the session cookie, or "".
func LoggedInUsername(db *sql.DB, r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	username, err := storage.GetSessionUsername(db, HashSessionToken(cookie.Value))
	if err != nil {
		return ""
	}
	return username
}
