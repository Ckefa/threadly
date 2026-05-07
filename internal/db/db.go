package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error
	env := os.Getenv("ENV")
	dbPath := os.Getenv("DB_PATH")
	dbHost := os.Getenv("DB_HOST")

	// Use SQLite for development if PostgreSQL is not available
	if env == "dev" && dbPath != "" {
		log.Println("Using development sqlite Database")
		DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	} else if dbHost != "" {
		log.Println("Using PostgreSQL database")

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			dbHost,
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		log.Fatal("Missing Database configuration: DB_PATH | DB_HOST")
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
}
