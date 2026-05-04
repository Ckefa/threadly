package handlers

import (
	"strconv"

	"threadly/internal/db"
	"threadly/internal/models"

	"github.com/gin-gonic/gin"
)

func ShowCustomers(c *gin.Context) {
	userID := c.GetUint("user_id")

	var clients []models.Client
	if err := db.DB.Where("user_id = ?", userID).Find(&clients).Error; err != nil {
		c.HTML(500, "index.html", gin.H{
			"Title": "Threadly",
			"Error": "Failed to load clients",
		})
		return
	}

	// Load conversation progress for each client
	type ClientWithProgress struct {
		models.Client
		Progress models.ConversationProgress `json:"progress"`
	}

	var clientsWithProgress []ClientWithProgress
	for _, client := range clients {
		var progress models.ConversationProgress
		err := db.DB.Table("conversation_progresses").
			Joins("JOIN conversations ON conversation_progresses.conversation_id = conversations.id").
			Where("conversations.client_id = ?", client.ID).
			First(&progress).Error

		// Create default progress if not found
		if err != nil {
			progress = models.ConversationProgress{
				CurrentStage:  models.StageInitial,
				ProgressScore: 10,
			}
		}

		clientsWithProgress = append(clientsWithProgress, ClientWithProgress{
			Client:   client,
			Progress: progress,
		})
	}

	c.HTML(200, "index.html", gin.H{
		"Title":   "Threadly",
		"Clients": clientsWithProgress,
	})
}

func CreateCustomer(c *gin.Context) {
	userID := c.GetUint("user_id")

	name := c.PostForm("name")
	email := c.PostForm("email")
	phone := c.PostForm("phone")

	client := models.Client{
		UserID: userID,
		Name:   name,
		Email:  email,
		Phone:  phone,
		Status: models.StatusNew,
	}

	if err := db.DB.Create(&client).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create client"})
		return
	}

	// Create conversation for the new client
	conversation := models.Conversation{
		ClientID: client.ID,
	}
	db.DB.Create(&conversation)

	c.JSON(200, gin.H{"client": client})
}

func UpdateCustomerStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	customerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid customer ID"})
		return
	}

	var customer models.Client
	if err := db.DB.Where("id = ? AND user_id = ?", customerID, userID).First(&customer).Error; err != nil {
		c.JSON(404, gin.H{"error": "Customer not found"})
		return
	}

	newStatus := c.PostForm("status")
	customer.Status = models.ClientStatus(newStatus)

	if err := db.DB.Save(&customer).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update customer status"})
		return
	}

	c.JSON(200, gin.H{"customer": customer})
}

func GetCustomerConversationID(c *gin.Context) {
	userID := c.GetUint("user_id")
	customerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid customer ID"})
		return
	}

	// Verify customer belongs to user
	var customer models.Client
	if err := db.DB.Where("id = ? AND user_id = ?", customerID, userID).First(&customer).Error; err != nil {
		c.JSON(404, gin.H{"error": "Customer not found"})
		return
	}

	// Get conversation for this customer
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", customerID).First(&conversation).Error; err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(200, gin.H{"conversation_id": conversation.ID})
}
