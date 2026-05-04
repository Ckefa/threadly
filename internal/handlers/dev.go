package handlers

import (
	"threadly/internal/db"
	"threadly/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

func DevPage(c *gin.Context) {
	// Get user info from context
	userID, exists := c.Get("user_id")
	var isLoggedIn bool
	if exists {
		isLoggedIn = userID != nil
	}

	// Get clients for display
	var clients []models.Client
	if exists {
		db.DB.Where("user_id = ?", userID).Find(&clients)
	}

	c.HTML(200, "test.html", gin.H{
		"Title":    "Dev Panel",
		"LoggedIn": isLoggedIn,
		"Clients":  clients,
	})
}

func Ping(c *gin.Context) {
	c.HTML(200, "ping.html", gin.H{
		"Status": "OK",
		"Time":   time.Now().Format("15:04:05"),
	})
}

func ListItems(c *gin.Context) {
	userID := c.GetUint("user_id")
	var clients []models.Client
	db.DB.Where("user_id = ?", userID).Find(&clients)
	c.HTML(200, "items.html", gin.H{
		"Items": clients,
		"Count": len(clients),
	})
}

func CreateItem(c *gin.Context) {
	userID := c.GetUint("user_id")
	name := c.PostForm("name")
	client := models.Client{
		UserID: userID,
		Name:   name,
		Status: models.StatusNew,
	}
	db.DB.Create(&client)
	ListItems(c)
}

func DeleteItem(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")
	db.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Client{})
	ListItems(c)
}
