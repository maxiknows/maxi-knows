package storage

import "database/sql"

// CompleteRequiredPasswordChange saves the new password hash
// and marks the required password change as completed.
func CompleteRequiredPasswordChange(db *sql.DB, userID int, newPasswordHash string) (bool, error) {
	result, err := db.Exec(`
		UPDATE users
		SET password = ?,
		    must_change_password = 0
		WHERE id = ?
		  AND must_change_password = 1
	`, newPasswordHash, userID)

	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows == 1, nil
}
