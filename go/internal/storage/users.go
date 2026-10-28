package storage

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("username or email already exists")
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
}

// GetUserByUsername returns ErrUserNotFound if no such user exists.
func GetUserByUsername(db *sql.DB, username string) (User, error) {
	var u User
	err := db.QueryRow(
		"SELECT id, username, password FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	return u, err
}

// UserExists reports whether the username or the email is already taken.
func UserExists(db *sql.DB, username, email string) (bool, error) {
	var id int
	err := db.QueryRow(
		"SELECT id FROM users WHERE username = ? OR email = ?",
		username, email,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateUser inserts a user. If the table has UNIQUE constraints on
// username/email, a violation is reported as ErrUserExists.
func CreateUser(db *sql.DB, username, email, passwordHash string) error {
	_, err := db.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, passwordHash,
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrUserExists
	}
	return err
}

func UpdatePassword(db *sql.DB, userID int, passwordHash string) error {
	_, err := db.Exec(
		"UPDATE users SET password = ? WHERE id = ?",
		passwordHash, userID,
	)
	return err
}
