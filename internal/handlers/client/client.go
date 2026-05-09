package client

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"threadly/internal/db"
	"threadly/internal/models"
)

func CreateClient(c *gin.Context) {
	businessID := c.GetUint("business_id")

	name := c.PostForm("name")
	email := c.PostForm("email")
	phone := c.PostForm("phone")

	client := models.Client{
		BusinessID: businessID,
		Name:       name,
		Email:      email,
		Phone:      phone,
		Status:     models.StatusNew,
	}

	if err := db.DB.Create(&client).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create client"})
		return
	}

	// Create conversation for the new client
	conversation := models.Conversation{
		ClientID:   client.ID,
		BusinessID: businessID,
	}
	db.DB.Create(&conversation)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Customer created successfully",
		"client":  client,
	})
}

func DeleteClient(c *gin.Context) {
	businessID := c.GetUint("business_id")
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid customer ID"})
		return
	}

	// Check if customer belongs to this user
	var client models.Client
	if err := db.DB.Where("id = ? AND business_id = ?", clientID, businessID).First(&client).Error; err != nil {
		c.JSON(404, gin.H{"error": "Customer not found"})
		return
	}

	// Start transaction
	tx := db.DB.Begin()

	// Delete conversation
	if err := tx.Where("client_id = ?", client.ID).Delete(&models.Conversation{}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to delete conversation"})
		return
	}

	// Delete customer
	if err := tx.Delete(&client).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to delete customer"})
		return
	}

	tx.Commit()

	c.JSON(200, gin.H{
		"success": true,
		"message": "Customer deleted successfully",
	})
}

func UpdateClientStatus(c *gin.Context) {
	businessID := c.GetUint("business_id")
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid customer ID"})
		return
	}

	var client models.Client
	if err := db.DB.Where("id = ? AND business_id = ?", clientID, businessID).First(&client).Error; err != nil {
		c.JSON(404, gin.H{"error": "Customer not found"})
		return
	}

	newStatus := c.PostForm("status")
	client.Status = models.ClientStatus(newStatus)

	if err := db.DB.Save(&client).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update customer status"})
		return
	}

	c.JSON(200, gin.H{"client": client})
}

func GetClientConversationID(c *gin.Context) {
	businessID := c.GetUint("business_id")
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid customer ID"})
		return
	}

	// Verify client belongs to business
	var client models.Client
	if err := db.DB.Where("id = ? AND business_id = ?", clientID, businessID).First(&client).Error; err != nil {
		c.JSON(404, gin.H{"error": "Customer not found"})
		return
	}

	// Get conversation for this customer
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", clientID).First(&conversation).Error; err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(200, gin.H{"conversation_id": conversation.ID})
}
