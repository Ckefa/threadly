package business

import (
	"fmt"
	"log"
	"net/http"
	"threadly/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateOrder  Creation
func (h *BusinessHandler) CreateOrder(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	var request struct {
		ClientID        uint   `json:"client_id" binding:"required"`
		ProductID       uint   `json:"product_id" binding:"required"`
		Quantity        int    `json:"quantity" binding:"required"`
		DeliveryAddress string `json:"delivery_address"`
		Notes           string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get product details
	var product models.Product
	if err := h.db.Where("id = ? AND business_id = ?", request.ProductID, businessID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check stock availability
	if product.Stock < request.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
		return
	}

	// Create or get client
	var client models.Client
	if err := h.db.Where("id = ?", request.ClientID).First(&client).Error; err != nil {
		log.Println("Failed to create client", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}

	// Create order
	order := models.Order{
		BusinessID:   businessID,
		ClientID:     client.ID,
		Quantity:     request.Quantity,
		OrderNumber:  generateOrderNumber(),
		Status:       "pending",
		Sender:       "business",
		TotalAmount:  float64(request.Quantity) * product.Price,
		Notes:        fmt.Sprintf("Delivery: %s. %s", request.DeliveryAddress, request.Notes),
		DeliveryDate: &[]time.Time{time.Now().AddDate(0, 0, 7)}[0],
	}

	if err := h.db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Create order item
	orderItem := models.OrderItem{
		OrderID:    order.ID,
		ProductID:  product.ID,
		Quantity:   request.Quantity,
		UnitPrice:  product.Price,
		TotalPrice: float64(request.Quantity) * product.Price,
	}

	if err := h.db.Create(&orderItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order item"})
		return
	}

	// Update product stock
	product.Stock -= request.Quantity
	if err := h.db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
		return
	}

	// Create inventory log
	inventoryLog := models.InventoryLog{
		ProductID: product.ID,
		Type:      "out",
		Quantity:  request.Quantity,
		Reason:    fmt.Sprintf("Order #%s", order.OrderNumber),
	}
	h.db.Create(&inventoryLog)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"order":   order,
		"message": fmt.Sprintf("Order %s created successfully", order.OrderNumber),
	})
}

func (h *BusinessHandler) GetOrders(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	var orders []models.Order
	h.db.Where("business_id = ?", businessID).Find(&orders)

	var pendingCount, confirmedCount, fulfilledCount, canceledCount int64
	var totalRevenue float64

	for _, order := range orders {
		switch order.Status {
		case "pending":
			pendingCount++
		case "confirmed":
			confirmedCount++
		case "fulfilled":
			fulfilledCount++
		case "canceled":
			canceledCount++
		}
		totalRevenue += order.TotalAmount
	}

	c.HTML(http.StatusOK, "orders.html", gin.H{
		"Business":       gin.H{},
		"Orders":         orders,
		"PendingCount":   pendingCount,
		"ConfirmedCount": confirmedCount,
		"FulfilledCount": fulfilledCount,
		"CanceledCount":  canceledCount,
		"TotalOrders":    len(orders),
		"TotalRevenue":   totalRevenue,
	})
}

func (h *BusinessHandler) UpdateOrderStatus(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	id := c.Param("id")
	var request struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := h.db.Where("id = ? AND business_id = ?", id, businessID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	order.Status = request.Status
	if err := h.db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "order": order})
}

// ClientCreateOrder allows customers to create orders
func (h *BusinessHandler) ClientCreateOrder(c *gin.Context) {
	// Get client ID from authenticated context (set by ClientMiddleware)
	clientID := c.GetUint("client_id")
	if clientID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated as client"})
		return
	}

	var request struct {
		BusinessID      uint   `json:"business_id" binding:"required"`
		ProductID       uint   `json:"product_id" binding:"required"`
		Quantity        int    `json:"quantity" binding:"required,min=1"`
		DeliveryAddress string `json:"delivery_address"`
		Notes           string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get product details
	var product models.Product
	if err := h.db.First(&product, request.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Get client by primary key
	var client models.Client
	if err := h.db.First(&client, clientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find client"})
		return
	}

	// Create order
	totalAmount := product.Price * float64(request.Quantity)
	order := models.Order{
		BusinessID:  request.BusinessID,
		ClientID:    client.ID,
		OrderNumber: generateOrderNumber(),
		Status:      "pending",
		Sender:      "client",
		Quantity:    request.Quantity,
		TotalAmount: totalAmount,
		Notes:       request.Notes,
	}

	if err := h.db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Create order item
	orderItem := models.OrderItem{
		OrderID:    order.ID,
		ProductID:  request.ProductID,
		Quantity:   request.Quantity,
		UnitPrice:  product.Price,
		TotalPrice: totalAmount,
	}

	if err := h.db.Create(&orderItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "order": order, "product_name": product.Name, "quantity": request.Quantity})
}

func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}
