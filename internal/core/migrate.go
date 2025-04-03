package core

import (
	"eventual/internal/config"
	"eventual/internal/db"
	"log"

	"github.com/joho/godotenv"
)

func Migrate() error {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Printf("[WARNING] %s", err)
	}

	dbConfig, err := config.ValidateEnvStruct(&config.Config{})
	if err != nil {
		return err
	}

	log.Printf("Migrating on: %v\n", dbConfig.DBPath)

	dbconn, err := db.NewDBConnection(dbConfig.DBPath)
	if err != nil {
		return err
	}

	err = db.Migrate(dbconn, true)
	return err
}
