// internal/database/database.go
package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func New(dbPath string) (*sql.DB, error) {
	// Ensure the directory exists
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		return nil, err
	}

	// Open the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Create the user_roles table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_roles (
			user_id TEXT,
			guild_id TEXT,
			roles TEXT,
			PRIMARY KEY (user_id, guild_id)
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
