package main

import (
	"log"
	"os"
	"strings"
	"text/template"

	"threadly/internal/db"
	"threadly/internal/models"
	"threadly/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	gin.SetMode(os.Getenv("GIN_MODE"))

	db.Connect()

	log.Println("🔄 Starting database auto-migration...")
	db.DB.AutoMigrate(
		&models.Business{},
		&models.Client{},
		&models.Conversation{},
		&models.Message{},
		&models.Action{},
		&models.ClientAuth{},
		&models.ConversationProgress{},
		&models.Product{},
		&models.Service{},
		&models.Order{},
		&models.OrderItem{},
		&models.Booking{},
		&models.BookingItem{},
		&models.Payment{},
		&models.InventoryLog{},
	)
	if err := db.MigrateNamingConventions(); err != nil {
		log.Fatalf("failed to apply naming migration: %v", err)
	}
	log.Println("✅ Database auto-migration completed successfully")

	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatalf("failed to set trusted proxies: %v", err)
	}

	// Add template functions
	r.SetFuncMap(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	})

	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./web/static")

	routes.Setup(r)

	log.Println("🚀 Running on :" + os.Getenv("APP_PORT"))
	r.Run(":" + os.Getenv("APP_PORT"))
}
