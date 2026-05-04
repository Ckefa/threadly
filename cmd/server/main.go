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
	)

	r := gin.Default()
	r.LoadHTMLFiles(
		"web/templates/index.html",
		"web/templates/login.html",
		"web/templates/register.html",
		"web/templates/chat.html",
		"web/templates/message_partial.html",
		"web/templates/customer_login.html",
		"web/templates/customer_otp.html",
		"web/templates/customer_dashboard.html",
		"web/templates/customer_chat.html",
		"web/templates/customer_message_partial.html",
		"web/templates/action_modal.html",
		"web/templates/test.html",
		"web/templates/partials/items.html",
		"web/templates/partials/ping.html",
	)
	r.Static("/static", "./web/static")

	routes.Setup(r)

	log.Println("🚀 Running on :" + os.Getenv("APP_PORT"))
	r.Run(":" + os.Getenv("APP_PORT"))
}
