package db

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error

	// Use SQLite for development if PostgreSQL is not available
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_HOST") == "localhost" {
		log.Println("Using SQLite for development")
		DB, err = gorm.Open(sqlite.Open("threadly.db"), &gorm.Config{})
	} else {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	// Log database type and file for debugging
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_HOST") == "localhost" {
		log.Println("Using SQLite database: threadly.db")
	} else {
		log.Println("Using PostgreSQL database")
	}
}

// MigrateNamingConventions renames legacy *_id columns to current naming.
func MigrateNamingConventions() error {
	tableColumns := map[string][]string{
		"clients":  {"user_id:business_id"},
		"orders":   {"user_id:business_id"},
		"bookings": {"user_id:business_id"},
		"products": {"user_id:business_id"},
		"services": {"user_id:business_id"},
	}

	for table, renames := range tableColumns {
		for _, pair := range renames {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid rename pair: %s", pair)
			}
			from, to := parts[0], parts[1]

			hasFrom := DB.Migrator().HasColumn(table, from)
			hasTo := DB.Migrator().HasColumn(table, to)
			if hasFrom && !hasTo {
				log.Printf("Renaming column %s.%s -> %s", table, from, to)
				if err := DB.Migrator().RenameColumn(table, from, to); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
