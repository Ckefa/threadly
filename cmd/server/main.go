package main

import (
	"log"
	"os"

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
	db.DB.AutoMigrate(
		&models.User{},
		&models.Client{},
		&models.Conversation{},
		&models.Message{},
		&models.Action{},
		&models.CustomerAuth{},
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

	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./web/static")

	routes.Setup(r)

	log.Println("🚀 Running on :" + os.Getenv("APP_PORT"))
	r.Run(":" + os.Getenv("APP_PORT"))
}
