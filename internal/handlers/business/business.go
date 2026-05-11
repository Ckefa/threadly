package business

import (
	"fmt"
	"net/http"
	"threadly/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BusinessHandler struct {
	db *gorm.DB
}

func NewBusinessHandler(db *gorm.DB) *BusinessHandler {
	return &BusinessHandler{db: db}
}

// DashboardData structure
type DashboardData struct {
	Business            models.Business
	ProductCount        int64
	ServiceCount        int64
	PendingOrderCount   int64
	PendingBookingCount int64
	TotalRevenue        float64
	TotalOrders         int64
	TotalBookings       int64
	ActiveClients       int64
	RecentOrders        []models.Order
	RecentBookings      []models.Booking
	LowStockProducts    []models.Product
}

func (h *BusinessHandler) GetBizHome(c *gin.Context) {
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

	if err := h.db.Raw(query, businessID, businessID).Scan(&clientsWithUnread).Error; err != nil {
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
	h.db.Model(&models.Order{}).Where("business_id = ? AND status = ?", businessID, "pending").Count(&pendingOrderCount)

	var pendingBookingCount int64
	h.db.Model(&models.Booking{}).Where("business_id = ? AND status = ?", businessID, "pending").Count(&pendingBookingCount)

	totalPending := int(pendingOrderCount + pendingBookingCount)

	c.HTML(200, "business.html", gin.H{
		"Title":               "Threadly",
		"Clients":             clientsWithUnread,
		"PendingOrderCount":   int(pendingOrderCount),
		"PendingBookingCount": int(pendingBookingCount),
		"TotalPending":        totalPending,
	})
}

func (h *BusinessHandler) GetDashboard(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.HTML(http.StatusUnauthorized, "business_login.html", gin.H{"error": "Business not authenticated"})
		return
	}

	// Get user from database
	var currentBusiness models.Business
	if err := h.db.First(&currentBusiness, businessID).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "business_login.html", gin.H{"error": "Business not found"})
		return
	}

	// Get counts
	var productCount, serviceCount, pendingOrderCount, pendingBookingCount int64
	var totalRevenue float64
	var totalOrders, totalBookings, activeClients int64

	h.db.Model(&models.Product{}).Where("business_id = ? AND is_active = ?", businessID, true).Count(&productCount)
	h.db.Model(&models.Service{}).Where("business_id = ? AND is_active = ?", businessID, true).Count(&serviceCount)
	h.db.Model(&models.Order{}).Where("business_id = ? AND status = ?", businessID, "pending").Count(&pendingOrderCount)
	h.db.Model(&models.Booking{}).Where("business_id = ? AND status = ?", businessID, "pending").Count(&pendingBookingCount)
	h.db.Model(&models.Order{}).Where("business_id = ?", businessID).Count(&totalOrders)
	h.db.Model(&models.Booking{}).Where("business_id = ?", businessID).Count(&totalBookings)
	h.db.Model(&models.Client{}).Where("business_id = ?", businessID).Count(&activeClients)

	// Calculate total revenue from completed orders and bookings
	var ordersRevenue, bookingsRevenue float64
	h.db.Model(&models.Order{}).Select("COALESCE(SUM(total_amount), 0)").Where("business_id = ? AND status IN ?", businessID, []string{"confirmed", "fulfilled"}).Scan(&ordersRevenue)
	h.db.Model(&models.Booking{}).Select("COALESCE(SUM(total_amount), 0)").Where("business_id = ? AND status IN ?", businessID, []string{"confirmed", "fulfilled"}).Scan(&bookingsRevenue)
	totalRevenue = ordersRevenue + bookingsRevenue

	// Get recent orders with client info
	var recentOrders []models.Order
	h.db.Preload("Client").Where("business_id = ?", businessID).Order("created_at DESC").Limit(5).Find(&recentOrders)

	// Get recent bookings with client info
	var recentBookings []models.Booking
	h.db.Preload("Client").Where("business_id = ?", businessID).Order("created_at DESC").Limit(5).Find(&recentBookings)

	// Get low stock products
	var lowStockProducts []models.Product
	h.db.Where("business_id = ? AND stock <= min_stock AND is_active = ?", businessID, true).Find(&lowStockProducts)

	data := DashboardData{
		Business:            currentBusiness,
		ProductCount:        productCount,
		ServiceCount:        serviceCount,
		PendingOrderCount:   pendingOrderCount,
		PendingBookingCount: pendingBookingCount,
		TotalRevenue:        totalRevenue,
		TotalOrders:         totalOrders,
		TotalBookings:       totalBookings,
		ActiveClients:       activeClients,
		RecentOrders:        recentOrders,
		RecentBookings:      recentBookings,
		LowStockProducts:    lowStockProducts,
	}

	c.HTML(http.StatusOK, "dashboard.html", data)
}

// Helper function to get or create conversation by client and business ID
func (h *BusinessHandler) getOrCreateConversation(clientID uint, businessID uint) (*models.Conversation, error) {
	var conversation models.Conversation
	err := h.db.Where("client_id = ? AND business_id = ?", clientID, businessID).First(&conversation).Error
	if err != nil {
		conversation = models.Conversation{
			ClientID:   clientID,
			BusinessID: businessID,
		}
		if err := h.db.Create(&conversation).Error; err != nil {
			return nil, fmt.Errorf("failed to create conversation: %v", err)
		}
		fmt.Printf("DEBUG BusinessHandler: Created new conversation ID=%d for client_id=%d, business_id=%d\n",
			conversation.ID, clientID, businessID)
	}
	return &conversation, nil
}

// Helper function to get or create client
func (h *BusinessHandler) getOrCreateClient(email string, businessID uint) (*models.Client, error) {
	var client models.Client
	err := h.db.Where("email = ? OR client_id = ?", email, businessID).First(&client).Error
	if err != nil {
		client = models.Client{
			ID:    businessID,
			Email: email,
			Name:  "Test client",
		}
		if err := h.db.Create(&client).Error; err != nil {
			return nil, fmt.Errorf("failed to create client: %v", err)
		}
		fmt.Printf("Created new client ID=%d for email=%s", client.ID, email)
	}
	return &client, nil
}
