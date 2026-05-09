package client

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"threadly/internal/db"
	"threadly/internal/models"
	"threadly/internal/services"

	"github.com/gin-gonic/gin"
)

func ShowClientLogin(c *gin.Context) {
	c.HTML(200, "client_login.html", gin.H{
		"Title": "Client Login - Threadly",
	})
}

func SendClientOTP(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.HTML(400, "client_login.html", gin.H{
			"Title": "Client Login - Threadly",
			"Error": "Email is required",
		})
		return
	}

	otp, err := services.SendClientOTP(email)
	if err != nil {
		c.HTML(400, "client_login.html", gin.H{
			"Title": "Client Login - Threadly",
			"Error": "Client not found",
		})
		return
	}

	c.HTML(200, "client_otp.html", gin.H{
		"Title": "Enter OTP - Threadly",
		"Email": email,
		"OTP":   otp, // For testing only
	})
}

func VerifyClientOTP(c *gin.Context) {
	email := c.PostForm("email")
	otpCode := c.PostForm("otp")

	if email == "" || otpCode == "" {
		c.HTML(400, "client_otp.html", gin.H{
			"Title": "Enter OTP - Threadly",
			"Email": email,
			"Error": "Email and OTP are required",
		})
		return
	}

	clientAuth, err := services.VerifyClientOTP(email, otpCode)
	if err != nil {
		c.HTML(400, "client_otp.html", gin.H{
			"Title": "Enter OTP - Threadly",
			"Email": email,
			"Error": "Invalid or expired OTP",
		})
		return
	}

	// Mark as verified
	clientAuth.IsVerified = true
	clientAuth.OTPCode = "" // Clear OTP after verification
	db.DB.Save(&clientAuth)

	// Update client online status
	now := time.Now()
	db.DB.Model(&models.Client{}).Where("id = ?", clientAuth.ClientID).Updates(map[string]interface{}{
		"is_online":    true,
		"last_seen_at": &now,
	})

	// Generate JWT token
	token, err := services.GenerateClientToken(clientAuth)
	if err != nil {
		c.HTML(500, "client_otp.html", gin.H{
			"Title": "Enter OTP - Threadly",
			"Email": email,
			"Error": "Failed to generate token",
		})
		return
	}

	// Set cookie and redirect
	c.SetCookie("client_token", token, 86400, "/", "", false, true)
	c.Redirect(http.StatusFound, "/client")
}

func ClientDashboard(c *gin.Context) {
	// Get client info from token
	token := c.GetHeader("Authorization")
	if token == "" {
		token, _ = c.Cookie("client_token")
	}

	if token == "" {
		c.Redirect(http.StatusFound, "/client/login")
		return
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	claims, err := services.ValidateToken(token)
	if err != nil || claims.Subject != "client" {
		c.Redirect(http.StatusFound, "/client/login")
		return
	}

	log.Printf("ClientDashboard: Loading businesses for email=%s", claims.Email)

	// Step 1: Get all clients for this client email
	var client models.Client
	db.DB.Where("email = ?", claims.Email).First(&client)
	log.Printf("ClientDashboard: Found %v client email=%s", client.Name, claims.Email)

	// Step 2: Get all conversations for these clients to find business IDs
	var conversations []models.Conversation
	db.DB.Where("client_id = ?", client.ID).Find(&conversations)
	log.Printf("ClientDashboard: Found %d conversations for client ID=%v", len(conversations), client.ID)

	// Extract unique business IDs from conversations
	var businessIDs []uint
	seen := make(map[uint]bool)
	for _, conv := range conversations {
		if conv.BusinessID > 0 && !seen[conv.BusinessID] {
			businessIDs = append(businessIDs, conv.BusinessID)
			seen[conv.BusinessID] = true
		}
	}
	log.Printf("ClientDashboard: Found %d unique business IDs: %v", len(businessIDs), businessIDs)

	var businesses []models.Business
	if len(businessIDs) > 0 {
		if err = db.DB.Where("id IN ?", businessIDs).Find(&businesses).Error; err != nil {
			log.Printf("ClientDashboard: Error fetching businesses: %v", err)
			c.HTML(500, "client.html", gin.H{
				"Title": "Client Dashboard - Threadly",
				"Error": "Failed to load businesses",
			})
			return
		}
	}

	log.Printf("ClientDashboard: Found %d businesses for email=%s", len(businesses), claims.Email)
	for i, b := range businesses {
		log.Printf("ClientDashboard: Business[%d] ID=%d, Name=%s, Type=%s", i, b.ID, b.FirstName, b.BusinessType)
	}

	c.HTML(200, "client.html", gin.H{
		"Title":      "Client Dashboard - Threadly",
		"Email":      claims.Email,
		"Businesses": businesses,
	})
}

type MessageObj struct {
	ID        uint        `json:"id"`
	MsgType   string      `json:"msgtype"` // "message", "order", "booking"
	Value     string      `json:"value"`   // string content for normal messages, empty for orders/bookings
	Data      interface{} `json:"data"`    // order object or booking object as JSON, null for normal messages
	Sender    string      `json:"sender"`
	CreatedAt time.Time   `json:"created_at"`
}

// Helper function to get or create conversation by client email and business ID
func getOrCreateConversation(clientID uint, businessID uint) (*models.Conversation, *models.Client, error) {

	if clientID == 0 || businessID == 0 {
		return nil, nil, fmt.Errorf("missing client_id or business_id field")
	}
	// Get client
	var client models.Client
	if err := db.DB.Where("id = ?", clientID).First(&client).Error; err != nil {
		log.Print("Client not found by id ", clientID)
	}

	// Get or create conversation by client_id AND business_id
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ? AND business_id = ?", clientID, businessID).First(&conversation).Error; err != nil {
		conversation = models.Conversation{
			ClientID:   clientID,
			BusinessID: businessID,
		}
		if err := db.DB.Create(&conversation).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to create conversation: %v", err)
		}
		log.Printf("Created new conversation ID=%d for client_id=%d, business_id=%d",
			conversation.ID, clientID, businessID)
	}

	log.Printf("Using conversation ID=%d (client_id=%d, business_id=%d)",
		conversation.ID, conversation.ClientID, conversation.BusinessID)

	return &conversation, &client, nil
}

func GetClientMessages(c *gin.Context) {
	clientID := c.GetUint("client_id")

	businessIDStr := c.Param("business_id")
	var businessID uint
	if _, err := fmt.Sscanf(businessIDStr, "%d", &businessID); err != nil {
		c.String(400, "Invalid business ID")
		return
	}

	log.Printf("GetClientMessages: clientID=%d, businessID=%d", clientID, businessID)

	// Get or create conversation using helper
	conversation, client, err := getOrCreateConversation(clientID, businessID)
	if err != nil {
		log.Printf("Error getting conversation: %v", err)
		c.String(500, "Failed to get conversation")
		return
	}

	// Get messages for this conversation
	var messages []models.Message
	if err := db.DB.Where("conversation_id = ?", conversation.ID).Order("created_at ASC").Find(&messages).Error; err != nil {
		c.String(500, "Failed to load messages")
		return
	}

	log.Printf("Loaded %d messages for conversation_id=%d", len(messages), conversation.ID)

	// Convert messages to MessageObj
	var messageObjs []MessageObj
	for _, msg := range messages {
		messageObj := MessageObj{
			ID:        msg.ID,
			MsgType:   "message",
			Value:     msg.Content,
			Data:      msg,
			Sender:    msg.Sender,
			CreatedAt: msg.CreatedAt,
		}
		messageObjs = append(messageObjs, messageObj)
		log.Printf("Added message ID=%d, Content='%s', Sender='%s', ConvoID=%d",
			msg.ID, msg.Content, msg.Sender, msg.ConversationID)
	}

	// Fetch orders
	var orders []models.Order
	db.DB.Where("client_id = ? AND business_id = ?", client.ID, businessID).Order("created_at ASC").Find(&orders)
	log.Printf("Found %d orders for client_id=%d, business_id=%d", len(orders), client.ID, businessID)

	for _, order := range orders {
		var orderItems []models.OrderItem
		db.DB.Where("order_id = ?", order.ID).Find(&orderItems)

		var productNames []string
		for _, item := range orderItems {
			var product models.Product
			db.DB.First(&product, item.ProductID)
			productNames = append(productNames, product.Name)
		}

		orderData := map[string]interface{}{
			"id":            order.ID,
			"order_number":  order.OrderNumber,
			"status":        order.Status,
			"quantity":      order.Quantity,
			"total_amount":  order.TotalAmount,
			"notes":         order.Notes,
			"created_at":    order.CreatedAt,
			"product_names": productNames,
		}

		messageObjs = append(messageObjs, MessageObj{
			ID:        order.ID + 10000,
			MsgType:   "order",
			Value:     "",
			Data:      orderData,
			Sender:    order.Sender,
			CreatedAt: order.CreatedAt,
		})
		log.Printf("Added order ID=%d to MessageObj", order.ID)
	}

	// Fetch bookings
	var bookings []models.Booking
	db.DB.Where("client_id = ? AND business_id = ?", client.ID, businessID).Order("created_at ASC").Find(&bookings)
	log.Printf("Found %d bookings for client_id=%d, business_id=%d", len(bookings), client.ID, businessID)

	for _, booking := range bookings {
		var bookingItems []models.BookingItem
		db.DB.Where("booking_id = ?", booking.ID).Find(&bookingItems)

		var serviceNames []string
		for _, item := range bookingItems {
			var service models.Service
			db.DB.First(&service, item.ServiceID)
			serviceNames = append(serviceNames, service.Name)
		}

		bookingData := map[string]interface{}{
			"id":             booking.ID,
			"booking_number": booking.BookingNumber,
			"status":         booking.Status,
			"scheduled_date": booking.ScheduledDate.Format("Jan 2, 2006 3:04 PM"),
			"duration":       booking.Duration,
			"total_amount":   booking.TotalAmount,
			"notes":          booking.Notes,
			"created_at":     booking.CreatedAt,
			"service_names":  serviceNames,
		}

		messageObjs = append(messageObjs, MessageObj{
			ID:        booking.ID + 20000,
			MsgType:   "booking",
			Value:     "",
			Data:      bookingData,
			Sender:    booking.Sender,
			CreatedAt: booking.CreatedAt,
		})
		log.Printf("Added booking ID=%d to MessageObj", booking.ID)
	}

	// Sort all messageObjs by CreatedAt
	for i := 0; i < len(messageObjs); i++ {
		for j := i + 1; j < len(messageObjs); j++ {
			if messageObjs[i].CreatedAt.After(messageObjs[j].CreatedAt) {
				messageObjs[i], messageObjs[j] = messageObjs[j], messageObjs[i]
			}
		}
	}

	log.Printf("Total MessageObjs: %d (messages=%d, orders=%d, bookings=%d)",
		len(messageObjs), len(messages), len(orders), len(bookings))

	// Get business info
	var business struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
		Type string `json:"business_type"`
	}
	db.DB.Raw("SELECT id, first_name || ' ' || last_name as name, business_type FROM businesses WHERE id = ?", businessID).First(&business)

	c.HTML(200, "client_chat.html", gin.H{
		"Business":    business,
		"Client":      client,
		"Messages":    messages,
		"MessageObjs": messageObjs,
	})
}

func CreateClientMessage(c *gin.Context) {
	clientID := c.GetUint("client_id")

	businessIDStr := c.Param("business_id")
	var businessID uint
	if _, err := fmt.Sscanf(businessIDStr, "%d", &businessID); err != nil {
		c.String(400, "Invalid business ID")
		return
	}

	log.Printf("CreateClientMessage: clientID=%d, businessID=%d", clientID, businessID)

	// Get or create conversation using helper (same as GetClientMessages)
	conversation, _, err := getOrCreateConversation(clientID, businessID)
	if err != nil {
		log.Printf("Error getting conversation: %v", err)
		c.String(500, "Failed to get conversation")
		return
	}

	content := c.PostForm("content")
	sender := c.PostForm("sender")
	if sender == "" {
		sender = "client"
	}

	// Create message with the correct conversation ID
	message := models.Message{
		ConversationID: conversation.ID,
		Content:        content,
		Type:           "message",
		Sender:         sender,
	}

	if err := db.DB.Create(&message).Error; err != nil {
		log.Printf("Error creating message: %v", err)
		c.String(500, "Failed to create message")
		return
	}

	log.Printf("Message created: ID=%d, ConvoID=%d, Content='%s', Sender='%s'",
		message.ID, message.ConversationID, message.Content, message.Sender)

	// Return the newly created message as MessageObj for HTMX
	messageObj := MessageObj{
		ID:        message.ID,
		MsgType:   "message",
		Value:     message.Content,
		Data:      nil,
		Sender:    message.Sender,
		CreatedAt: message.CreatedAt,
	}

	c.HTML(200, "client_message_partial.html", gin.H{
		"MessageObj": messageObj,
	})
}

func ClientMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token, _ = c.Cookie("client_token")
		}

		if token == "" {
			c.Redirect(http.StatusFound, "/client/login")
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := services.ValidateToken(token)
		if err != nil || claims.Subject != "client" {
			c.Redirect(http.StatusFound, "/client/login")
			c.Abort()
			return
		}

		c.Set("client_id", claims.UserID)
		c.Set("client_email", claims.Email)
		c.Next()
	}
}

func ClientHeartbeat(c *gin.Context) {
	// Get client info from token
	token := c.GetHeader("Authorization")
	if token == "" {
		token, _ = c.Cookie("client_token")
	}

	if token == "" {
		c.JSON(401, gin.H{"error": "No token"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	claims, err := services.ValidateToken(token)
	if err != nil || claims.Subject != "client" {
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	// Update client online status
	now := time.Now()
	db.DB.Model(&models.Client{}).Where("id = ?", claims.UserID).Updates(map[string]interface{}{
		"is_online":    true,
		"last_seen_at": &now,
	})

	c.JSON(200, gin.H{"status": "ok", "timestamp": now})
}

func ClientLogout(c *gin.Context) {
	// Get client info from token
	token, _ := c.Cookie("client_token")
	if token != "" {
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := services.ValidateToken(token)
		if err == nil && claims.Subject == "client" {
			// Update client offline status
			db.DB.Model(&models.Client{}).Where("id = ?", claims.UserID).Update("is_online", false)
		}
	}

	// Clear cookie and redirect
	c.SetCookie("client_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/client/login")
}

// ClientUpdateOrder allows clients to update their order notes and quantity
func ClientUpdateOrder(c *gin.Context) {
	clientID := c.GetUint("client_id")
	orderID := c.Param("id")

	var order models.Order
	if err := db.DB.Where("id = ? AND client_id = ?", orderID, clientID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Update notes if provided
	notes := c.PostForm("notes")
	if notes != "" {
		order.Notes = notes
	}

	// Update quantity if provided
	quantityStr := c.PostForm("quantity")
	if quantityStr != "" {
		if quantity, err := strconv.Atoi(quantityStr); err == nil && quantity > 0 {
			// Get the order item to find product price
			var orderItem models.OrderItem
			if err := db.DB.Where("order_id = ?", order.ID).First(&orderItem).Error; err == nil {
				// Recalculate total amount
				order.TotalAmount = float64(quantity) * orderItem.UnitPrice
				orderItem.Quantity = quantity
				orderItem.TotalPrice = order.TotalAmount
				db.DB.Save(&orderItem)
			}
			// Update the main order quantity field
			order.Quantity = quantity
		}
	}

	db.DB.Save(&order)
	c.JSON(http.StatusOK, gin.H{"success": true, "order": order})
}

// ClientUpdateBooking allows clients to update their booking notes and date
func ClientUpdateBooking(c *gin.Context) {
	clientID := c.GetUint("client_id")
	bookingID := c.Param("id")

	var booking models.Booking
	if err := db.DB.Where("id = ? AND client_id = ?", bookingID, clientID).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	notes := c.PostForm("notes")
	scheduledDate := c.PostForm("scheduled_date")

	if notes != "" {
		booking.Notes = notes
	}
	if scheduledDate != "" {
		if newDate, err := time.Parse(time.RFC3339, scheduledDate); err == nil {
			booking.ScheduledDate = newDate
		}
	}

	db.DB.Save(&booking)
	c.JSON(http.StatusOK, gin.H{"success": true, "booking": booking})
}
