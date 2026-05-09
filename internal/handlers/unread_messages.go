package handlers

import (
	"threadly/internal/db"
	"threadly/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

func ShowClientsWithUnread(c *gin.Context) {
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
		c.HTML(500, "business.html", gin.H{
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

	c.HTML(200, "business.html", gin.H{
		"Title":               "Threadly",
		"Clients":             clientsWithUnread,
		"PendingOrderCount":   int(pendingOrderCount),
		"PendingBookingCount": int(pendingBookingCount),
		"TotalPending":        totalPending,
	})
}

func ShowBusinesses(c *gin.Context) {
	clientID := c.GetUint("client_id")

	// Business with unread count struct
	type BusinessWithUnread struct {
		models.Business
		ConversationID uint       `json:"conversation_id"`
		UnreadCount    int        `json:"unread_count"`
		LastMessageAt  *time.Time `json:"last_message_at"`
		OnlineStatus   string     `json:"online_status"`
		IsOnline       bool       `json:"is_online"`
	}

	var businessesWithUnread []BusinessWithUnread

	// Query: join businesses with their conversations, count unread messages
	query := `
		SELECT 
			businesses.*, 
			conversations.id as conversation_id,
			COUNT(CASE WHEN messages.sender = 'client' AND messages.created_at > COALESCE(conversations.last_read_by_client_at, '1970-01-01') THEN 1 END) as unread_count,
			MAX(messages.created_at) as last_message_at
		FROM businesses 
		JOIN conversations ON conversations.client_id = businesses.id AND conversations.business_id = businesses.id
		LEFT JOIN messages ON messages.conversation_id = conversations.id
		WHERE businesses.id = ?
		GROUP BY businesses.id, conversations.id
		ORDER BY unread_count DESC, last_message_at DESC
	`

	if err := db.DB.Raw(query, clientID).Scan(&businessesWithUnread).Error; err != nil {
		c.HTML(500, "dashboard.html", gin.H{
			"Title": "Threadly",
			"Error": "Failed to load businesses",
		})
		return
	}

	// Set online status for each business
	for i := range businessesWithUnread {
		if businessesWithUnread[i].IsOnline {
			businessesWithUnread[i].OnlineStatus = "online"
		} else {
			businessesWithUnread[i].OnlineStatus = "offline"
		}
	}

	c.HTML(200, "dashboard.html", gin.H{
		"Title":      "Client Dashboard - Threadly",
		"Businesses": businessesWithUnread,
	})
}
