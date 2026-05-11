package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"threadly/internal/db"
	"threadly/internal/models"

	"github.com/gin-gonic/gin"
)

type MessageObj struct {
	ID        uint        `json:"id"`
	MsgType   string      `json:"msgtype"` // "message", "order", "booking"
	Value     string      `json:"value"`   // string content for normal messages, empty for orders/bookings
	Data      interface{} `json:"data"`    // order object or booking object as JSON, null for normal messages
	Sender    string      `json:"sender"`
	CreatedAt time.Time   `json:"created_at"`
}

func GetMessages(c *gin.Context) {
	businessID := c.GetUint("business_id")
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Println("GetMessages: =>> Invalid customer ID")
		c.String(400, "Invalid customer ID")
		return
	}

	log.Println("Getting messages bizID, clientID", businessID, clientID)

	// Verify client belongs to business
	var client models.Client
	if err := db.DB.Where("id = ? AND business_id = ?", clientID, businessID).First(&client).Error; err != nil {
		log.Println("GetMessages: =>> Customer not found by id", clientID)
		c.String(404, "Customer not found")
		return
	}

	// Get conversation
	var conversation models.Conversation
	if err := db.DB.Where("client_id = ?", clientID).First(&conversation).Error; err != nil {
		log.Println("GetMessages: =>> Conversation not found", clientID, businessID)
		c.String(404, "Conversation not found")
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
			log.Println("GetMessages: =>> Failed to Crete conversation progress", clientID, businessID)
			c.String(500, "Failed to create conversation progress")
			return
		}
	}

	// Convert messages to MessageObj
	var messageObjs []MessageObj
	var messages []models.Message
	db.DB.Where("conversation_id = ?", conversation.ID).Order("created_at ASC").Find(&messages)

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
		var serviceName string
		var bookingItems []models.BookingItem
		db.DB.Where("booking_id = ?", booking.ID).Find(&bookingItems)

		for _, item := range bookingItems {
			var service models.Service
			if err := db.DB.First(&service, item.ServiceID).Error; err == nil {
				serviceName = service.Name
				break
			}
		}

		bookingData := map[string]interface{}{
			"id":             booking.ID,
			"service_name":   serviceName,
			"scheduled_date": booking.ScheduledDate,
			"notes":          booking.Notes,
			"status":         booking.Status,
			"created_at":     booking.CreatedAt,
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

	// Sort messageObjs by CreatedAt
	for i := 0; i < len(messageObjs); i++ {
		for j := i + 1; j < len(messageObjs); j++ {
			if messageObjs[i].CreatedAt.After(messageObjs[j].CreatedAt) {
				messageObjs[i], messageObjs[j] = messageObjs[j], messageObjs[i]
			}
		}
	}

	// Debug logging
	fmt.Printf("Loading chat for client %d, conversation ID: %d\n", clientID, conversation.ID)
	fmt.Printf("Progress data: %+v\n", progress)

	c.HTML(200, "business_chat.html", gin.H{
		"Customer": client,
		"Messages": messageObjs,
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

func MarkConversationAsRead(c *gin.Context) {
	businessID := c.GetUint("business_id")
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid client ID"})
		return
	}

	// Update conversation's last read time
	now := time.Now()
	if err := db.DB.Model(&models.Conversation{}).
		Where("business_id = ? AND client_id = ?", businessID, clientID).
		Update("last_read_by_business_at", &now).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to mark conversation as read"})
		return
	}

	// Also mark all unread messages as read by business
	if err := db.DB.Model(&models.Message{}).
		Where("conversation_id IN (SELECT id FROM conversations WHERE business_id = ? AND client_id = ?) AND sender = 'client' AND read_by_business = ?", businessID, clientID, false).
		Updates(map[string]interface{}{
			"read_by_business": true,
			"read_at":          &now,
		}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to mark messages as read"})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}

func MarkClientConversationAsRead(c *gin.Context) {
	clientID := c.GetUint("client_id")
	businessID, err := strconv.ParseUint(c.Param("business_id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid business ID"})
		return
	}

	now := time.Now()
	if err := db.DB.Model(&models.Conversation{}).
		Where("client_id = ? AND business_id = ?", clientID, businessID).
		Update("last_read_by_client_at", &now).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to mark conversation as read"})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}
