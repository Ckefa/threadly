package handlers

import (
	"strconv"
	"time"

	"threadly/internal/db"
	"threadly/internal/models"

	"github.com/gin-gonic/gin"
)

func Showclients(c *gin.Context) {
	businessID := c.GetUint("business_id")

	// Client with unread count struct
	type ClientWithUnread struct {
		models.Client
		ConversationID uint       `json:"conversation_id"`
		UnreadCount    int        `json:"unread_count"`
		LastMessageAt  *time.Time `json:"last_message_at"`
		OnlineStatus   string     `json:"online_status"`
	}

	var clientsWithUnread []ClientWithUnread

	// Query: join clients with their conversations, count unread messages
	query := `
		SELECT 
			clients.*, 
			conversations.id as conversation_id,
			COUNT(CASE WHEN messages.sender = 'client' AND messages.created_at > COALESCE(conversations.last_read_by_business_at, '1970-01-01') THEN 1 END) as unread_count,
			MAX(messages.created_at) as last_message_at
		FROM clients 
		JOIN conversations ON conversations.client_id = clients.id AND conversations.business_id = ?
		LEFT JOIN messages ON messages.conversation_id = conversations.id
		WHERE clients.business_id = ?
		GROUP BY clients.id, conversations.id
		ORDER BY unread_count DESC, last_message_at DESC
	`

	if err := db.DB.Raw(query, businessID, businessID).Scan(&clientsWithUnread).Error; err != nil {
		c.HTML(500, "index.html", gin.H{
			"Title": "Threadly",
			"Error": "Failed to load clients",
		})
		return
	}

	// Set online status for each client
	for i := range clientsWithUnread {
		if clientsWithUnread[i].IsOnline {
			clientsWithUnread[i].OnlineStatus = "online"
		} else {
			clientsWithUnread[i].OnlineStatus = "offline"
		}
	}

	// Count pending orders and bookings
	var pendingOrderCount int64
	db.DB.Model(&models.Order{}).Where("business_id = ? AND status = ?", businessID, "pending").Count(&pendingOrderCount)

	var pendingBookingCount int64
	db.DB.Model(&models.Booking{}).Where("business_id = ? AND status = ?", businessID, "pending").Count(&pendingBookingCount)

	totalPending := int(pendingOrderCount + pendingBookingCount)

	c.HTML(200, "index.html", gin.H{
		"Title":               "Threadly",
		"Clients":             clientsWithUnread,
		"PendingOrderCount":   int(pendingOrderCount),
		"PendingBookingCount": int(pendingBookingCount),
		"TotalPending":        totalPending,
	})
}

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
