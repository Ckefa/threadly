package routes

import (
	"threadly/internal/db"
	"threadly/internal/handlers"
	"threadly/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	// Initialize business handler
	businessHandler := handlers.NewBusinessHandler(db.DB)

	// PUBLIC - Business Routes
	r.GET("/login", handlers.ShowLogin)
	r.GET("/register", handlers.ShowRegister)
	r.POST("/login", handlers.Login)
	r.POST("/register", handlers.Register)
	r.GET("/logout", handlers.Logout)

	// PUBLIC - client Routes
	r.GET("/client/login", handlers.ShowClientLogin)
	r.POST("/client/send-otp", handlers.SendClientOTP)
	r.POST("/client/verify-otp", handlers.VerifyClientOTP)
	r.GET("/client/logout", handlers.ClientLogout)

	// PROTECTED BUSINESS ROUTES
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		// Chat Dashboard (original)
		protected.GET("/", handlers.Showclients)

		// Business Dashboard routes
		protected.GET("/business", businessHandler.GetDashboard)
		protected.GET("/products", businessHandler.GetProducts)
		protected.GET("/products/:id", businessHandler.GetProduct)
		protected.POST("/products", businessHandler.CreateProduct)
		protected.PUT("/products/:id", businessHandler.UpdateProduct)
		protected.DELETE("/products/:id", businessHandler.DeleteProduct)
		protected.GET("/services", businessHandler.GetServices)
		protected.GET("/services/:id", businessHandler.GetService)
		protected.POST("/services", businessHandler.CreateService)
		protected.PUT("/services/:id", businessHandler.UpdateService)
		protected.DELETE("/services/:id", businessHandler.DeleteService)
		protected.GET("/orders", businessHandler.GetOrders)
		protected.POST("/orders", businessHandler.CreateOrder)
		protected.PUT("/orders/:id/status", businessHandler.UpdateOrderStatus)
		protected.GET("/bookings", businessHandler.GetBookings)
		protected.GET("/bookings/:id", businessHandler.GetBooking)
		protected.POST("/bookings", businessHandler.CreateBooking)
		protected.PUT("/bookings/:id", businessHandler.UpdateBooking)
		protected.PUT("/bookings/:id/status", businessHandler.UpdateBookingStatus)

		// client routes
		protected.POST("/clients", handlers.CreateClient)
		protected.DELETE("/clients/:id", handlers.DeleteClient)
		protected.PUT("/clients/:id/status", handlers.UpdateClientStatus)
		protected.GET("/clients/:id/conversation-id", handlers.GetClientConversationID)

		// Message routes
		protected.GET("/clients/:id/messages", handlers.GetMessages)
		protected.POST("/clients/:id/messages", handlers.CreateMessage)
		protected.PUT("/messages/:message_id", handlers.UpdateMessage)

		// Action routes
		protected.POST("/messages/:message_id/actions", handlers.CreateAction)
		protected.POST("/messages/:message_id/actions/enhanced", handlers.CreateActionWithProgress)
		protected.GET("/actions", handlers.GetActions)
		protected.PUT("/actions/:id/status", handlers.UpdateActionStatus)

		// Conversation status route
		protected.PUT("/conversations/:conversation_id/status", handlers.UpdateConversationStatus)

		// Enhanced action routes
		protected.GET("/actions/modal/:message_id", handlers.ShowActionModal)
		protected.POST("/actions/enhanced", handlers.CreateEnhancedAction)
		protected.GET("/actions/enhanced", handlers.GetEnhancedActions)
		protected.PUT("/actions/:id/enhanced-status", handlers.UpdateEnhancedActionStatus)

		// Business widget routes
		protected.POST("/clients/:id/quick-booking", handlers.QuickBooking)
		protected.POST("/clients/:id/quick-order", handlers.QuickOrder)
		protected.POST("/clients/:id/request-payment", handlers.RequestPayment)
		protected.POST("/clients/:id/set-goal", handlers.SetGoal)

		// Conversation progress routes
		protected.GET("/conversations/:conversation_id/progress", handlers.GetConversationProgress)
		protected.PUT("/conversations/:conversation_id/stage", handlers.UpdateConversationStage)
	}

	// PROTECTED client ROUTES
	clientProtected := r.Group("/client")
	clientProtected.Use(handlers.ClientMiddleware())
	{
		clientProtected.GET("/dashboard", handlers.ClientDashboard)
		clientProtected.GET("/businesses/:business_id/messages", handlers.GetClientMessages)
		clientProtected.POST("/businesses/:business_id/messages", handlers.CreateClientMessage)
		clientProtected.GET("/businesses/:business_id/products", businessHandler.GetBusinessProducts)
		clientProtected.GET("/businesses/:business_id/services", businessHandler.GetBusinessServices)
		clientProtected.POST("/businesses/:business_id/bookings", businessHandler.ClientCreateBooking)
		clientProtected.POST("/orders", businessHandler.ClientCreateOrder)
		clientProtected.POST("/bookings", businessHandler.ClientCreateBooking)
		clientProtected.POST("/orders/:id/update", handlers.ClientUpdateOrder)
		clientProtected.POST("/bookings/:id/update", handlers.ClientUpdateBooking)
		clientProtected.POST("/heartbeat", handlers.ClientHeartbeat)
	}

	// API
	api := r.Group("/api/v1")
	{
		api.GET("/ping", handlers.Ping)
	}

	// DEV ONLY
	if gin.Mode() == gin.DebugMode {
		dev := r.Group("/test")
		{
			dev.GET("", handlers.DevPage)
			dev.GET("/items", handlers.ListItems)
			dev.POST("/items", handlers.CreateItem)
			dev.DELETE("/items/:id", handlers.DeleteItem)
		}
	}
}
