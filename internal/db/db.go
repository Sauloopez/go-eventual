package db

import (
	"log"
	"path"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewDBConnection creates a new database connection.
func NewDBConnection(directory string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path.Join(directory, "events.db")), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to SQLite database in %s", directory)
	return db, nil
}
