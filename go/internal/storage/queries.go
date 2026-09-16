package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type Page struct {
	Title    string
	URL      string
	Language string
	Content  string
}

func main() {
	db, err := sql.Open("sqlite", "../data/whoknows.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// This inserts a hardcoded test user named "johndoe".
	// Because the username and email are UNIQUE in the database,
	// running this file more than once may cause a UNIQUE constraint error
	// if johndoe already exists.
	lastID, err := InsertUserQuery(db)
	if err != nil {
		log.Printf("InsertUserQuery error: %v", err)
	} else {
		fmt.Printf("InsertUserQuery: Inserted user with id %d\n", lastID)
	}

	userID, err := GetUserIDQuery(db)
	if err != nil {
		log.Printf("GetUserIDQuery error: %v", err)
	} else {
		fmt.Printf("GetUserIDQuery: User 'johndoe' has id %d\n", userID)
	}

	id, username, email, password, err := GetUserByIDQuery(db)
	if err != nil {
		log.Printf("GetUserByIDQuery error: %v", err)
	} else {
		fmt.Printf(
			"GetUserByIDQuery: id=%d username=%s email=%s password=%s\n",
			id,
			username,
			email,
			password,
		)
	}

	id, username, email, password, err = GetUserByUsernameQuery(db)
	if err != nil {
		log.Printf("GetUserByUsernameQuery error: %v", err)
	} else {
		fmt.Printf(
			"GetUserByUsernameQuery: id=%d username=%s email=%s password=%s\n",
			id,
			username,
			email,
			password,
		)
	}

	pages, err := SearchPagesQuery(db, "golang", "en")
	if err != nil {
		log.Printf("SearchPagesQuery error: %v", err)
	} else {
		for _, page := range pages {
			fmt.Printf(
				"SearchPagesQuery: title=%s url=%s language=%s content=%s\n",
				page.Title,
				page.URL,
				page.Language,
				page.Content,
			)
		}
	}
}

// InsertUserQuery inserts a hardcoded test user into the database.
// This is currently only used for testing the database connection and queries.
// It should later be replaced with a function that accepts real user input.
func InsertUserQuery(db *sql.DB) (int64, error) {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ('johndoe', 'john@example.com', '5f4dcc3b5aa765d61d8327deb882cf99')
	`

	res, err := db.Exec(query)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}

func GetUserIDQuery(db *sql.DB) (int, error) {
	query := "SELECT id FROM users WHERE username = 'johndoe'"

	var id int

	err := db.QueryRow(query).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetUserByIDQuery(db *sql.DB) (int, string, string, string, error) {
	query := "SELECT id, username, email, password FROM users WHERE id = 1"

	row := db.QueryRow(query)

	var id int
	var username, email, password string

	err := row.Scan(&id, &username, &email, &password)
	if err != nil {
		return 0, "", "", "", err
	}

	return id, username, email, password, nil
}

func GetUserByUsernameQuery(db *sql.DB) (int, string, string, string, error) {
	query := `
		SELECT id, username, email, password
		FROM users
		WHERE username = 'johndoe'
	`

	row := db.QueryRow(query)

	var id int
	var username, email, password string

	err := row.Scan(&id, &username, &email, &password)
	if err != nil {
		return 0, "", "", "", err
	}

	return id, username, email, password, nil
}

func SearchPagesQuery(db *sql.DB, searchTerm string, language string) ([]Page, error) {
	query := `
		SELECT title, url, language, content
		FROM pages
		WHERE language = ?
		AND content LIKE ?
	`

	rows, err := db.Query(query, language, "%"+searchTerm+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []Page

	for rows.Next() {
		var page Page

		err := rows.Scan(
			&page.Title,
			&page.URL,
			&page.Language,
			&page.Content,
		)

		if err != nil {
			log.Printf("SearchPagesQuery row scan error: %v", err)
			continue
		}

		pages = append(pages, page)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}
