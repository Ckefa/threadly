package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"threadly/internal/models"

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

func (h *BusinessHandler) GetDashboard(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Business not authenticated"})
		return
	}

	// Get user from database
	var currentBusiness models.Business
	if err := h.db.First(&currentBusiness, businessID).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Business not found"})
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

	c.HTML(http.StatusOK, "business_dashboard.html", data)
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

// GetProducts for business
func (h *BusinessHandler) GetProducts(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Business not authenticated"})
		return
	}

	var currentBusiness models.Business
	if err := h.db.First(&currentBusiness, businessID).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Business not found"})
		return
	}

	var products []models.Product
	h.db.Where("business_id = ?", businessID).Order("created_at DESC").Find(&products)

	c.HTML(http.StatusOK, "products.html", gin.H{
		"Business": currentBusiness,
		"Products": products,
	})
}

func (h *BusinessHandler) CreateProduct(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.BusinessID = businessID
	product.IsActive = true

	if err := h.db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "product": product})
}

func (h *BusinessHandler) GetProduct(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := h.db.Where("id = ? AND business_id = ?", productID, businessID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "product": product})
}

func (h *BusinessHandler) UpdateProduct(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := h.db.Where("id = ? AND business_id = ?", productID, businessID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "product": product})
}

func (h *BusinessHandler) DeleteProduct(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := h.db.Where("id = ? AND business_id = ?", productID, businessID).Delete(&models.Product{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetServices for the business
func (h *BusinessHandler) GetServices(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Business not authenticated"})
		return
	}

	var currentBusiness models.Business
	if err := h.db.First(&currentBusiness, businessID).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Business not found"})
		return
	}

	var services []models.Service
	h.db.Where("business_id = ?", businessID).Order("created_at DESC").Find(&services)

	c.HTML(http.StatusOK, "services.html", gin.H{
		"Business": currentBusiness,
		"Services": services,
	})
}

func (h *BusinessHandler) CreateService(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service.BusinessID = businessID
	service.IsActive = true

	if err := h.db.Create(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "service": service})
}

func (h *BusinessHandler) GetService(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var service models.Service
	if err := h.db.Where("id = ? AND business_id = ?", serviceID, businessID).First(&service).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "service": service})
}

func (h *BusinessHandler) UpdateService(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var service models.Service
	if err := h.db.Where("id = ? AND business_id = ?", serviceID, businessID).First(&service).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "service": service})
}

func (h *BusinessHandler) DeleteService(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	if err := h.db.Where("id = ? AND business_id = ?", serviceID, businessID).Delete(&models.Service{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

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
	if err := h.db.Where("client_id ? =", request.ClientID).First(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}

	// Create order
	order := models.Order{
		BusinessID:   businessID,
		ClientID:     client.ID,
		OrderNumber:  generateOrderNumber(),
		Status:       "pending",
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

func (h *BusinessHandler) GetBookings(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	var bookings []models.Booking
	h.db.Where("business_id = ?", businessID).Find(&bookings)

	var pendingCount, confirmedCount, completedCount, canceledCount int64
	var totalRevenue float64

	for _, booking := range bookings {
		switch booking.Status {
		case "pending":
			pendingCount++
		case "confirmed":
			confirmedCount++
		case "completed":
			completedCount++
		case "canceled":
			canceledCount++
		}
		totalRevenue += booking.TotalAmount
	}

	c.HTML(http.StatusOK, "bookings.html", gin.H{
		"Business":       gin.H{},
		"Bookings":       bookings,
		"PendingCount":   pendingCount,
		"ConfirmedCount": confirmedCount,
		"CompletedCount": completedCount,
		"CanceledCount":  canceledCount,
		"TotalBookings":  len(bookings),
		"TotalRevenue":   totalRevenue,
	})
}

func (h *BusinessHandler) GetBooking(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	bookingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	var booking models.Booking
	if err := h.db.Where("id = ? AND business_id = ?", bookingID, businessID).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "booking": booking})
}

func (h *BusinessHandler) UpdateBookingStatus(c *gin.Context) {
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

	var booking models.Booking
	if err := h.db.Where("id = ? AND business_id = ?", id, businessID).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	booking.Status = request.Status
	if err := h.db.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "booking": booking})
}

func (h *BusinessHandler) UpdateBooking(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	bookingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	var booking models.Booking
	if err := h.db.Where("id = ? AND business_id = ?", bookingID, businessID).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	var request struct {
		ServiceID   uint   `json:"service_id"`
		BookingDate string `json:"booking_date"`
		Notes       string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking.Notes = request.Notes

	if request.BookingDate != "" {
		scheduledDate, err := time.Parse(time.RFC3339, request.BookingDate)
		if err == nil {
			booking.ScheduledDate = scheduledDate
		}
	}

	if request.ServiceID > 0 {
		var service models.Service
		if err := h.db.First(&service, request.ServiceID).Error; err == nil {
			if len(booking.BookingItems) > 0 {
				booking.BookingItems[0].ServiceID = request.ServiceID
				booking.BookingItems[0].UnitPrice = service.MaxPrice
				booking.BookingItems[0].TotalPrice = service.MaxPrice
				h.db.Save(&booking.BookingItems[0])
			}
			booking.TotalAmount = service.MaxPrice
		}
	}

	if err := h.db.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "booking": booking})
}

// CreateBooking for business
func (h *BusinessHandler) CreateBooking(c *gin.Context) {
	businessID := c.GetUint("business_id")
	if businessID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Business not authenticated"})
		return
	}

	var request struct {
		ServiceID   uint   `json:"service_id" binding:"required"`
		clientName  string `json:"client_name" binding:"required"`
		clientEmail string `json:"client_email"`
		clientPhone string `json:"client_phone"`
		BookingDate string `json:"booking_date" binding:"required"`
		Notes       string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Booking binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get service details
	var service models.Service
	if err := h.db.Where("id = ? AND business_id = ?", request.ServiceID, businessID).First(&service).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// Create or get client
	var client models.Client
	if err := h.db.Where("client_email ? =", request.clientEmail).First(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}

	// Parse booking date
	bookingDate, err := parseBookingDateTime(request.BookingDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create booking
	booking := models.Booking{
		BusinessID:    businessID,
		ClientID:      client.ID,
		BookingNumber: generateBookingNumber(),
		Status:        "pending",
		ScheduledDate: bookingDate,
		Duration:      service.Duration,
		TotalAmount:   service.MaxPrice,
		Notes:         request.Notes,
	}

	if err := h.db.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	// Create booking item
	bookingItem := models.BookingItem{
		BookingID:  booking.ID,
		ServiceID:  service.ID,
		UnitPrice:  service.MaxPrice,
		TotalPrice: service.MaxPrice,
	}

	if err := h.db.Create(&bookingItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"booking": booking,
		"message": fmt.Sprintf("Booking %s created successfully", booking.BookingNumber),
	})
}

// GetBusinessProducts as a struct
func (h *BusinessHandler) GetBusinessProducts(c *gin.Context) {
	businessID, err := strconv.ParseUint(c.Param("business_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid business ID"})
		return
	}

	var products []models.Product
	if err := h.db.Where("business_id = ? AND is_active = ?", businessID, true).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *BusinessHandler) GetBusinessServices(c *gin.Context) {
	businessID, err := strconv.ParseUint(c.Param("business_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid business ID"})
		return
	}

	var services []models.Service
	if err := h.db.Where("business_id = ? AND is_active = ?", businessID, true).Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch services"})
		return
	}

	c.JSON(http.StatusOK, services)
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

// ClientCreateBooking allows customers to create bookings
func (h *BusinessHandler) ClientCreateBooking(c *gin.Context) {
	// Get client ID from authenticated context (set by ClientMiddleware)
	clientID := c.GetUint("client_id")
	if clientID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated as client"})
		return
	}

	var request struct {
		ServiceID     uint   `json:"service_id" binding:"required"`
		ScheduledDate string `json:"scheduled_date" binding:"required"`
		Notes         string `json:"notes"`
		BusinessID    uint   `json:"business_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get service details
	var service models.Service
	if err := h.db.First(&service, request.ServiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// Get client by primary key
	var client models.Client
	if err := h.db.First(&client, clientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find client"})
		return
	}

	// Parse booking date
	bookingDate, err := time.Parse(time.RFC3339, request.ScheduledDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// Create booking
	booking := models.Booking{
		BusinessID:    request.BusinessID,
		ClientID:      client.ID,
		BookingNumber: generateBookingNumber(),
		Status:        "pending",
		ScheduledDate: bookingDate,
		Duration:      service.Duration,
		TotalAmount:   service.MaxPrice,
		Notes:         request.Notes,
	}

	if err := h.db.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking!"})
		return
	}

	// Create booking item
	bookingItem := models.BookingItem{
		BookingID:  booking.ID,
		ServiceID:  request.ServiceID,
		Quantity:   1,
		UnitPrice:  service.MaxPrice,
		TotalPrice: service.MaxPrice,
	}

	if err := h.db.Create(&bookingItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "booking": booking, "service_name": service.Name})
}

// Helper functions
func parseBookingDateTime(bookingDateTime string) (time.Time, error) {
	parts := strings.Split(bookingDateTime, "T")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid datetime format, expected dateTtime")
	}

	date, err := time.Parse("2006-01-02", parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %v", err)
	}

	timeOnly, err := time.Parse("15:04", parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time: %v", err)
	}

	result := time.Date(date.Year(), date.Month(), date.Day(),
		timeOnly.Hour(), timeOnly.Minute(), 0, 0, time.UTC)

	return result, nil
}

func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}

func generateBookingNumber() string {
	return fmt.Sprintf("BOOK-%d", time.Now().Unix())
}
