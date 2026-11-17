package storage

import "database/sql"

// MigratePasswordReset adds a password reset requirement
// for all existing users after the security breach.
func MigratePasswordReset(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check whether the column already exists.
	rows, err := tx.Query("PRAGMA table_info(users)")
	if err != nil {
		return err
	}

	columnExists := false

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int

		err = rows.Scan(
			&cid,
			&name,
			&columnType,
			&notNull,
			&defaultValue,
			&primaryKey,
		)
		if err != nil {
			rows.Close()
			return err
		}

		if name == "must_change_password" {
			columnExists = true
		}
	}

	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	// Only migrate once.
	if !columnExists {
		_, err = tx.Exec(`
			ALTER TABLE users
			ADD COLUMN must_change_password INTEGER NOT NULL DEFAULT 0
		`)
		if err != nil {
			return err
		}

		// Require all existing users to change password.
		_, err = tx.Exec(`
			UPDATE users
			SET must_change_password = 1
		`)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
