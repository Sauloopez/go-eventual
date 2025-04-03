package db

import (
	"log"
	"path"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewDBConnection creates a new database connection.
func NewDBConnection(directory string) (*gorm.DB, error) {
	dbPath := path.Join(directory, "events.db")
	log.Printf("[LOG] Connecting to database in %s", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Printf("[LOG] Connected to SQLite database in %s", directory)
	return db, nil
}
