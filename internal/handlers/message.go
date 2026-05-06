package handlers

import (
	"fmt"
	"strconv"

	"threadly/internal/db"
	"threadly/internal/models"

	"github.com/gin-gonic/gin"
)

func GetMessages(c *gin.Context) {
	businessID := c.GetUint("business_id")
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.String(400, "Invalid customer ID")
		return
	}

	// Verify client belongs to business
	var client models.Client
	if err := db.DB.Where("id = ? AND business_id = ?", clientID, businessID).First(&client).Error; err != nil {
		c.String(404, "Customer not found")
		return
	}

	// Get conversation
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", clientID).First(&conversation).Error; err != nil {
		c.String(404, "Conversation not found")
		return
	}

	// Get messages
	var messages []models.Message
	if err := db.DB.Where("conversation_id = ?", conversation.ID).Order("created_at ASC").Find(&messages).Error; err != nil {
		c.String(500, "Failed to load messages")
		return
	}

	// Add conversation ID to client struct for template use
	client.ConversationID = conversation.ID

	// Load conversation progress
	var progress models.ConversationProgress
	if err := db.DB.Where("conversation_id = ?", conversation.ID).First(&progress).Error; err != nil {
		// Create default progress if not exists
		progress = models.ConversationProgress{
			ConversationID: conversation.ID,
			CurrentStage:   models.StageInitial,
			ProgressScore:  10,
		}
		if err := db.DB.Create(&progress).Error; err != nil {
			c.String(500, "Failed to create conversation progress")
			return
		}
	}

	// Debug logging
	fmt.Printf("Loading chat for client %d, conversation ID: %d\n", clientID, conversation.ID)
	fmt.Printf("Progress data: %+v\n", progress)

	c.HTML(200, "chat.html", gin.H{
		"Customer": client,
		"Messages": messages,
		"Progress": progress,
	})
}

func CreateMessage(c *gin.Context) {
	businessID := c.GetUint("business_id")
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.String(400, "Invalid customer ID")
		return
	}

	// Verify client belongs to business
	var client models.Client
	if err := db.DB.Where("id = ? AND business_id = ?", clientID, businessID).First(&client).Error; err != nil {
		c.String(404, "Customer not found")
		return
	}

	// Get conversation
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", clientID).First(&conversation).Error; err != nil {
		c.String(404, "Conversation not found")
		return
	}

	content := c.PostForm("content")
	sender := c.PostForm("sender") // "user" or "client"

	message := models.Message{
		ConversationID: conversation.ID,
		Content:        content,
		Sender:         sender,
	}

	if err := db.DB.Create(&message).Error; err != nil {
		c.String(500, "Failed to create message")
		return
	}

	// Return message partial
	c.HTML(200, "message_partial.html", gin.H{
		"Message": message,
	})
}

func UpdateMessage(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("message_id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid message ID"})
		return
	}

	var request struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var message models.Message
	if err := db.DB.First(&message, messageID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Message not found"})
		return
	}

	message.Content = request.Content
	if err := db.DB.Save(&message).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": message})
}
