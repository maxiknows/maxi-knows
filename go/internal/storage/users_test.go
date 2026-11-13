
package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
	"errors"

	_ "modernc.org/sqlite"
)

// Create a temporary SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := InitDB(db); err != nil {
		t.Fatalf("Could not initialize test database: %v", err)
	}

	return db
}

func TestCreateAndGetUser(t *testing.T) {
	db := setupTestDB(t)

	// Create a user
	err := CreateUser(db, "testuser", "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("Could not create user: %v", err)
	}

	// Retrieve the user
	user, err := GetUserByUsername(db, "testuser")
	if err != nil {
		t.Fatalf("Could not retrieve user: %v", err)
	}

	// Verify the stored data
	if user.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", user.Username)
	}

	if user.PasswordHash != "hashedpassword" {
		t.Errorf("Expected password hash hashedpassword, got %s", user.PasswordHash)
	}
}


func TestCreateDuplicateUser(t *testing.T) {
	db := setupTestDB(t)

	// Create the first user
	err := CreateUser(db, "testuser", "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("Could not create first user: %v", err)
	}

	// Try to create a user with the same username
	err = CreateUser(db, "testuser", "another@example.com", "hashedpassword")

	// A duplicate username should return an error
	if !errors.Is(err, ErrUserExists) {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}


func TestCreateDuplicateEmail(t *testing.T) {
	db := setupTestDB(t)

	// Create the first user
	err := CreateUser(db, "testuser1", "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("Could not create first user: %v", err)
	}

	// Try to create another user with the same email
	err = CreateUser(db, "testuser2", "test@example.com", "hashedpassword")

	// A duplicate email should return an error
	if !errors.Is(err, ErrUserExists) {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}



func TestGetUserNotFound(t *testing.T) {
	db := setupTestDB(t)

	// Try to retrieve a user that does not exist
	_, err := GetUserByUsername(db, "nonexistent")

	// Expect ErrUserNotFound
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
