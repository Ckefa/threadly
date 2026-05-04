package handlers

import (
	"fmt"
	"strconv"

	"threadly/internal/db"
	"threadly/internal/models"

	"github.com/gin-gonic/gin"
)

func GetMessages(c *gin.Context) {
	userID := c.GetUint("user_id")
	customerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.String(400, "Invalid customer ID")
		return
	}

	// Verify customer belongs to user
	var customer models.Client
	if err := db.DB.Where("id = ? AND user_id = ?", customerID, userID).First(&customer).Error; err != nil {
		c.String(404, "Customer not found")
		return
	}

	// Get conversation
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", customerID).First(&conversation).Error; err != nil {
		c.String(404, "Conversation not found")
		return
	}

	// Get messages
	var messages []models.Message
	if err := db.DB.Where("conversation_id = ?", conversation.ID).Order("created_at ASC").Find(&messages).Error; err != nil {
		c.String(500, "Failed to load messages")
		return
	}

	// Add conversation ID to customer struct for template use
	customer.ConversationID = conversation.ID

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
	fmt.Printf("Loading chat for customer %d, conversation ID: %d\n", customerID, conversation.ID)
	fmt.Printf("Progress data: %+v\n", progress)

	c.HTML(200, "chat.html", gin.H{
		"Customer": customer,
		"Messages": messages,
		"Progress": progress,
	})
}

func CreateMessage(c *gin.Context) {
	userID := c.GetUint("user_id")
	customerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.String(400, "Invalid customer ID")
		return
	}

	// Verify customer belongs to user
	var customer models.Client
	if err := db.DB.Where("id = ? AND user_id = ?", customerID, userID).First(&customer).Error; err != nil {
		c.String(404, "Customer not found")
		return
	}

	// Get conversation
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", customerID).First(&conversation).Error; err != nil {
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
