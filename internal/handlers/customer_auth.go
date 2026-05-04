package handlers

import (
	"net/http"
	"strings"
	"time"

	"threadly/internal/db"
	"threadly/internal/models"
	"threadly/internal/services"

	"github.com/gin-gonic/gin"
)

func ShowCustomerLogin(c *gin.Context) {
	c.HTML(200, "customer_login.html", gin.H{
		"Title": "Customer Login - Threadly",
	})
}

func SendCustomerOTP(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.HTML(400, "customer_login.html", gin.H{
			"Title": "Customer Login - Threadly",
			"Error": "Email is required",
		})
		return
	}

	otp, err := services.SendCustomerOTP(email)
	if err != nil {
		c.HTML(400, "customer_login.html", gin.H{
			"Title": "Customer Login - Threadly",
			"Error": "Customer not found",
		})
		return
	}

	c.HTML(200, "customer_otp.html", gin.H{
		"Title": "Enter OTP - Threadly",
		"Email": email,
		"OTP":   otp, // For testing only
	})
}

func VerifyCustomerOTP(c *gin.Context) {
	email := c.PostForm("email")
	otpCode := c.PostForm("otp")

	if email == "" || otpCode == "" {
		c.HTML(400, "customer_otp.html", gin.H{
			"Title": "Enter OTP - Threadly",
			"Email": email,
			"Error": "Email and OTP are required",
		})
		return
	}

	customerAuth, err := services.VerifyCustomerOTP(email, otpCode)
	if err != nil {
		c.HTML(400, "customer_otp.html", gin.H{
			"Title": "Enter OTP - Threadly",
			"Email": email,
			"Error": "Invalid or expired OTP",
		})
		return
	}

	// Mark as verified
	customerAuth.IsVerified = true
	customerAuth.OTPCode = "" // Clear OTP after verification
	db.DB.Save(&customerAuth)

	// Update customer online status
	now := time.Now()
	db.DB.Model(&models.Client{}).Where("id = ?", customerAuth.ClientID).Updates(map[string]interface{}{
		"is_online":    true,
		"last_seen_at": &now,
	})

	// Generate JWT token
	token, err := services.GenerateCustomerToken(customerAuth)
	if err != nil {
		c.HTML(500, "customer_otp.html", gin.H{
			"Title": "Enter OTP - Threadly",
			"Email": email,
			"Error": "Failed to generate token",
		})
		return
	}

	// Set cookie and redirect
	c.SetCookie("customer_token", token, 86400, "/", "", false, true)
	c.Redirect(http.StatusFound, "/customer/dashboard")
}

func CustomerDashboard(c *gin.Context) {
	// Get customer info from token
	token := c.GetHeader("Authorization")
	if token == "" {
		token, _ = c.Cookie("customer_token")
	}

	if token == "" {
		c.Redirect(http.StatusFound, "/customer/login")
		return
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	claims, err := services.ValidateToken(token)
	if err != nil || claims.Subject != "customer" {
		c.Redirect(http.StatusFound, "/customer/login")
		return
	}

	// Get customer's businesses (the businesses they have conversations with)
	var businesses []struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Type  string `json:"business_type"`
	}

	err = db.DB.Raw(`
		SELECT DISTINCT u.id, u.first_name || ' ' || u.last_name as name, u.email, u.business_type
		FROM users u
		JOIN clients c ON u.id = c.user_id
		JOIN conversations conv ON c.id = conv.client_id
		WHERE c.email = ?
	`, claims.Email).Scan(&businesses).Error

	if err != nil {
		c.HTML(500, "customer_dashboard.html", gin.H{
			"Title": "Customer Dashboard - Threadly",
			"Error": "Failed to load businesses",
		})
		return
	}

	c.HTML(200, "customer_dashboard.html", gin.H{
		"Title":      "Customer Dashboard - Threadly",
		"Email":      claims.Email,
		"Businesses": businesses,
	})
}

func GetCustomerMessages(c *gin.Context) {
	customerEmail := c.GetString("customer_email")
	businessID := c.Param("business_id")

	// Get the customer's client record for this business
	var client models.Client
	err := db.DB.Raw(`
		SELECT c.* FROM clients c
		JOIN users u ON c.user_id = u.id
		WHERE c.email = ? AND u.id = ?
	`, customerEmail, businessID).First(&client).Error

	if err != nil {
		c.String(404, "Conversation not found")
		return
	}

	// Get conversation
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", client.ID).First(&conversation).Error; err != nil {
		c.String(404, "Conversation not found")
		return
	}

	// Get messages
	var messages []models.Message
	if err := db.DB.Where("conversation_id = ?", conversation.ID).Order("created_at ASC").Find(&messages).Error; err != nil {
		c.String(500, "Failed to load messages")
		return
	}

	// Get business info
	var business struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
		Type string `json:"business_type"`
	}
	db.DB.Raw("SELECT id, first_name || ' ' || last_name as name, business_type FROM users WHERE id = ?", businessID).First(&business)

	c.HTML(200, "customer_chat.html", gin.H{
		"Business": business,
		"Client":   client,
		"Messages": messages,
	})
}

func CreateCustomerMessage(c *gin.Context) {
	customerEmail := c.GetString("customer_email")
	businessID := c.Param("business_id")

	// Get the customer's client record for this business
	var client models.Client
	err := db.DB.Raw(`
		SELECT c.* FROM clients c
		JOIN users u ON c.user_id = u.id
		WHERE c.email = ? AND u.id = ?
	`, customerEmail, businessID).First(&client).Error

	if err != nil {
		c.String(404, "Conversation not found")
		return
	}

	// Get conversation
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", client.ID).First(&conversation).Error; err != nil {
		c.String(404, "Conversation not found")
		return
	}

	content := c.PostForm("content")
	sender := "client" // Customer messages are always from client

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
	c.HTML(200, "customer_message_partial.html", gin.H{
		"Message": message,
	})
}

func CustomerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token, _ = c.Cookie("customer_token")
		}

		if token == "" {
			c.Redirect(http.StatusFound, "/customer/login")
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := services.ValidateToken(token)
		if err != nil || claims.Subject != "customer" {
			c.Redirect(http.StatusFound, "/customer/login")
			c.Abort()
			return
		}

		c.Set("customer_id", claims.UserID)
		c.Set("customer_email", claims.Email)
		c.Next()
	}
}

func CustomerHeartbeat(c *gin.Context) {
	// Get customer info from token
	token := c.GetHeader("Authorization")
	if token == "" {
		token, _ = c.Cookie("customer_token")
	}

	if token == "" {
		c.JSON(401, gin.H{"error": "No token"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	claims, err := services.ValidateToken(token)
	if err != nil || claims.Subject != "customer" {
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	// Update customer online status
	now := time.Now()
	db.DB.Model(&models.Client{}).Where("id = ?", claims.UserID).Updates(map[string]interface{}{
		"is_online":    true,
		"last_seen_at": &now,
	})

	c.JSON(200, gin.H{"status": "ok", "timestamp": now})
}

func CustomerLogout(c *gin.Context) {
	// Get customer info from token
	token, _ := c.Cookie("customer_token")
	if token != "" {
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := services.ValidateToken(token)
		if err == nil && claims.Subject == "customer" {
			// Update customer offline status
			db.DB.Model(&models.Client{}).Where("id = ?", claims.UserID).Update("is_online", false)
		}
	}

	// Clear cookie and redirect
	c.SetCookie("customer_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/customer/login")
}
