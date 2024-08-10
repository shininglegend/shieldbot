// internal/database/database.go
package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func New(url string) (*sql.DB, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Create the user_roles table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_roles (
			user_id VARCHAR(255),
			guild_id VARCHAR(255),
			roles TEXT,
			PRIMARY KEY (user_id, guild_id)
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
