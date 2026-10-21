package storage

import (
	"database/sql"
	"time"
)

// CreateSession saves a new session in the database.
// Only the hash of the session token is stored.
func CreateSession(db *sql.DB, tokenHash string, userID int, expiresAt time.Time) error {
	_, err := db.Exec(
		`INSERT INTO sessions (token_hash, user_id, expires_at)
		 VALUES (?, ?, ?)`,
		tokenHash,
		userID,
		expiresAt.Unix(),
	)

	return err
}

 // GetSessionUsername returns the username belonging to
 // a valid, non-expired session.
 func GetSessionUsername(db *sql.DB, tokenHash string) (string, error) {
     var username string

     err := db.QueryRow(`
         SELECT users.username
         FROM sessions
         JOIN users ON sessions.user_id = users.id
         WHERE sessions.token_hash = ?
           AND sessions.expires_at > ?
     `, tokenHash, time.Now().Unix()).Scan(&username)

     return username, err
 }

 // DeleteSession removes a session from the database.
 func DeleteSession(db *sql.DB, tokenHash string) error {
     _, err := db.Exec(
         "DELETE FROM sessions WHERE token_hash = ?",
         tokenHash,
     )

     return err
 }