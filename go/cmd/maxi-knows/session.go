
package main

import (
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/hex"
    "net/http"

    "maxi-knows/internal/storage"
)

// generateSessionToken creates a random session token
// and a SHA-256 hash for database storage.
func generateSessionToken() (string, string, error) {
	tokenBytes := make([]byte, 32)

	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	return token, tokenHash, nil
}


 // getLoggedInUsername finds the user associated
 // with the session cookie.
 func getLoggedInUsername(db *sql.DB, r *http.Request) string {
     cookie, err := r.Cookie("session")
     if err != nil {
         return ""
     }

     // Hash the cookie value before looking it up in SQLite.
     hash := sha256.Sum256([]byte(cookie.Value))
     tokenHash := hex.EncodeToString(hash[:])

     username, err := storage.GetSessionUsername(db, tokenHash)
     if err != nil {
         return ""
     }

     return username
 }

 // hashSessionToken returns the SHA-256 hash of a session token.
func hashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
