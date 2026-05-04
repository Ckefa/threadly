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

// Dashboard data structure
type DashboardData struct {
	User                models.User
	ProductCount        int64
	ServiceCount        int64
	PendingOrderCount   int64
	PendingBookingCount int64
	TotalRevenue        float64
	TotalOrders         int64
	TotalBookings       int64
	ActiveCustomers     int64
	RecentOrders        []models.Order
	RecentBookings      []models.Booking
	LowStockProducts    []models.Product
}

func (h *BusinessHandler) GetDashboard(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "User not authenticated"})
		return
	}

	// Get user from database
	var currentUser models.User
	if err := h.db.First(&currentUser, userID).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "User not found"})
		return
	}

	// Get counts
	var productCount, serviceCount, pendingOrderCount, pendingBookingCount int64
	var totalRevenue float64
	var totalOrders, totalBookings, activeCustomers int64

	h.db.Model(&models.Product{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&productCount)
	h.db.Model(&models.Service{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&serviceCount)
	h.db.Model(&models.Order{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&pendingOrderCount)
	h.db.Model(&models.Booking{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&pendingBookingCount)
	h.db.Model(&models.Order{}).Where("user_id = ?", userID).Count(&totalOrders)
	h.db.Model(&models.Booking{}).Where("user_id = ?", userID).Count(&totalBookings)
	h.db.Model(&models.Client{}).Where("user_id = ?", userID).Count(&activeCustomers)

	// Calculate total revenue from completed orders and bookings
	var ordersRevenue, bookingsRevenue float64
	h.db.Model(&models.Order{}).Select("COALESCE(SUM(total_amount), 0)").Where("user_id = ? AND status IN ?", userID, []string{"confirmed", "fulfilled"}).Scan(&ordersRevenue)
	h.db.Model(&models.Booking{}).Select("COALESCE(SUM(total_amount), 0)").Where("user_id = ? AND status IN ?", userID, []string{"confirmed", "fulfilled"}).Scan(&bookingsRevenue)
	totalRevenue = ordersRevenue + bookingsRevenue

	// Get recent orders with customer info
	var recentOrders []models.Order
	h.db.Preload("Customer").Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&recentOrders)

	// Get recent bookings with customer info
	var recentBookings []models.Booking
	h.db.Preload("Customer").Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&recentBookings)

	// Get low stock products
	var lowStockProducts []models.Product
	h.db.Where("user_id = ? AND stock <= min_stock AND is_active = ?", userID, true).Find(&lowStockProducts)

	data := DashboardData{
		User:                currentUser,
		ProductCount:        productCount,
		ServiceCount:        serviceCount,
		PendingOrderCount:   pendingOrderCount,
		PendingBookingCount: pendingBookingCount,
		TotalRevenue:        totalRevenue,
		TotalOrders:         totalOrders,
		TotalBookings:       totalBookings,
		ActiveCustomers:     activeCustomers,
		RecentOrders:        recentOrders,
		RecentBookings:      recentBookings,
		LowStockProducts:    lowStockProducts,
	}

	c.HTML(http.StatusOK, "business_dashboard.html", data)
}

// Products Management
func (h *BusinessHandler) GetProducts(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "User not authenticated"})
		return
	}

	// Get user from database
	var currentUser models.User
	if err := h.db.First(&currentUser, userID).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "User not found"})
		return
	}

	var products []models.Product
	h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&products)

	c.HTML(http.StatusOK, "products.html", gin.H{
		"User":     currentUser,
		"Products": products,
	})
}

func (h *BusinessHandler) CreateProduct(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.UserID = userID
	product.IsActive = true

	if err := h.db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "product": product})
}

func (h *BusinessHandler) UpdateProduct(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
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
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).Delete(&models.Product{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Services Management
func (h *BusinessHandler) GetServices(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "User not authenticated"})
		return
	}

	// Get user from database
	var currentUser models.User
	if err := h.db.First(&currentUser, userID).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "User not found"})
		return
	}

	var services []models.Service
	h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&services)

	c.HTML(http.StatusOK, "services.html", gin.H{
		"User":     currentUser,
		"Services": services,
	})
}

func (h *BusinessHandler) CreateService(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service.UserID = userID
	service.IsActive = true

	if err := h.db.Create(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "service": service})
}

func (h *BusinessHandler) UpdateService(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var service models.Service
	if err := h.db.Where("id = ? AND user_id = ?", serviceID, userID).First(&service).Error; err != nil {
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
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	if err := h.db.Where("id = ? AND user_id = ?", serviceID, userID).Delete(&models.Service{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Quick Order Creation
func (h *BusinessHandler) CreateOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		ProductID       uint   `json:"product_id" binding:"required"`
		Quantity        int    `json:"quantity" binding:"required"`
		CustomerName    string `json:"customer_name" binding:"required"`
		CustomerEmail   string `json:"customer_email"`
		CustomerPhone   string `json:"customer_phone"`
		DeliveryAddress string `json:"delivery_address"`
		Notes           string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get product details
	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", request.ProductID, userID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check stock availability
	if product.Stock < request.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
		return
	}

	// Create or get customer
	var customer models.Client
	customerResult := h.db.Where("email = ? AND user_id = ?", request.CustomerEmail, userID).First(&customer)
	if customerResult.Error != nil {
		// Create new customer
		customer = models.Client{
			UserID: userID,
			Name:   request.CustomerName,
			Email:  request.CustomerEmail,
			Phone:  request.CustomerPhone,
			Status: "active",
		}
		if err := h.db.Create(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
			return
		}
	}

	// Create order
	order := models.Order{
		UserID:       userID,
		CustomerID:   customer.ID,
		OrderNumber:  generateOrderNumber(),
		Status:       "pending",
		TotalAmount:  float64(request.Quantity) * product.Price,
		Notes:        fmt.Sprintf("Delivery: %s. %s", request.DeliveryAddress, request.Notes),
		DeliveryDate: &[]time.Time{time.Now().AddDate(0, 0, 7)}[0], // Default 7 days delivery
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
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var orders []models.Order
	h.db.Where("user_id = ?", userID).Find(&orders)

	// Calculate stats
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
		"User":           gin.H{},
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
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
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
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	order.Status = request.Status

	if err := h.db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"order":   order,
	})
}

func (h *BusinessHandler) GetBookings(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var bookings []models.Booking
	h.db.Where("user_id = ?", userID).Find(&bookings)

	// Calculate stats
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
		"User":           gin.H{},
		"Bookings":       bookings,
		"PendingCount":   pendingCount,
		"ConfirmedCount": confirmedCount,
		"CompletedCount": completedCount,
		"CanceledCount":  canceledCount,
		"TotalBookings":  len(bookings),
		"TotalRevenue":   totalRevenue,
	})
}

func (h *BusinessHandler) UpdateBookingStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
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
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	booking.Status = request.Status

	if err := h.db.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"booking": booking,
	})
}

// Quick Booking Creation
func (h *BusinessHandler) CreateBooking(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		ServiceID     uint   `json:"service_id" binding:"required"`
		CustomerName  string `json:"customer_name" binding:"required"`
		CustomerEmail string `json:"customer_email"`
		CustomerPhone string `json:"customer_phone"`
		BookingDate   string `json:"booking_date" binding:"required"`
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Booking binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get service details
	var service models.Service
	if err := h.db.Where("id = ? AND user_id = ?", request.ServiceID, userID).First(&service).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// Create or get customer
	var customer models.Client
	customerResult := h.db.Where("email = ? AND user_id = ?", request.CustomerEmail, userID).First(&customer)
	if customerResult.Error != nil {
		// Create new customer
		customer = models.Client{
			UserID: userID,
			Name:   request.CustomerName,
			Email:  request.CustomerEmail,
			Phone:  request.CustomerPhone,
			Status: "active",
		}
		if err := h.db.Create(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
			return
		}
	}

	// Create booking
	// Parse booking date and time properly
	bookingDate, err := parseBookingDateTime(request.BookingDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking := models.Booking{
		UserID:        userID,
		CustomerID:    customer.ID,
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

// Helper function to properly parse booking datetime
func parseBookingDateTime(bookingDateTime string) (time.Time, error) {
	parts := strings.Split(bookingDateTime, "T")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid datetime format, expected dateTtime")
	}

	// Parse date
	date, err := time.Parse("2006-01-02", parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %v", err)
	}

	// Parse time
	timeOnly, err := time.Parse("15:04", parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time: %v", err)
	}

	// Combine date and time into proper datetime
	result := time.Date(date.Year(), date.Month(), date.Day(),
		timeOnly.Hour(), timeOnly.Minute(), 0, 0, time.UTC)

	return result, nil
}

// Helper functions
func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}

func generateBookingNumber() string {
	return fmt.Sprintf("BOOK-%d", time.Now().Unix())
}
