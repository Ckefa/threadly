package main

import (
	"html/template"
	"log"
	"os"
	"strings"
	"threadly/internal/db"
	"threadly/internal/models"
	"threadly/internal/routes"
	"time"

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
	log.Println("✅ Database auto-migration completed successfully")

	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatalf("failed to set trusted proxies: %v", err)
	}
	// Add template functions
	r.SetFuncMap(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					key := values[i].(string)
					dict[key] = values[i+1]
				}
			}
			return dict, nil
		},
		"title": strings.Title,
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" {
				return def
			}
			return val
		},
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
		"formatTime": func(t time.Time) string {
			return t.Format("3:04 PM")
		},
	})
	// Load templates from multiple directories
	r.LoadHTMLGlob("web/templates/**/**/*.html")

	// Get Static Path
	r.Static("/static", "./web/static")

	routes.Setup(r)
	routes.SetupClientRoutes(r)
	routes.SetupBusinessRoutes(r)

	log.Println("🚀 Running on :" + os.Getenv("APP_PORT"))
	r.Run(":" + os.Getenv("APP_PORT"))
}
