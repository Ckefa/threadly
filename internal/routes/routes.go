package routes

import (
	"threadly/internal/handlers"
	"threadly/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {

	// PUBLIC - Business Routes
	r.GET("/login", handlers.ShowLogin)
	r.GET("/register", handlers.ShowRegister)
	r.POST("/login", handlers.Login)
	r.POST("/register", handlers.Register)
	r.GET("/logout", handlers.Logout)

	// PUBLIC - Customer Routes
	r.GET("/customer/login", handlers.ShowCustomerLogin)
	r.POST("/customer/send-otp", handlers.SendCustomerOTP)
	r.POST("/customer/verify-otp", handlers.VerifyCustomerOTP)
	r.GET("/customer/logout", handlers.CustomerLogout)

	// PROTECTED BUSINESS ROUTES
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/", handlers.ShowCustomers)

		// Customer routes
		protected.POST("/customers", handlers.CreateCustomer)
		protected.PUT("/customers/:id/status", handlers.UpdateCustomerStatus)
		protected.GET("/customers/:id/conversation-id", handlers.GetCustomerConversationID)

		// Message routes
		protected.GET("/customers/:id/messages", handlers.GetMessages)
		protected.POST("/customers/:id/messages", handlers.CreateMessage)

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
		protected.POST("/customers/:id/quick-booking", handlers.QuickBooking)
		protected.POST("/customers/:id/quick-order", handlers.QuickOrder)
		protected.POST("/customers/:id/request-payment", handlers.RequestPayment)
		protected.POST("/customers/:id/set-goal", handlers.SetGoal)

		// Conversation progress routes
		protected.GET("/conversations/:conversation_id/progress", handlers.GetConversationProgress)
		protected.PUT("/conversations/:conversation_id/stage", handlers.UpdateConversationStage)
	}

	// PROTECTED CUSTOMER ROUTES
	customerProtected := r.Group("/customer")
	customerProtected.Use(handlers.CustomerMiddleware())
	{
		customerProtected.GET("/dashboard", handlers.CustomerDashboard)
		customerProtected.GET("/businesses/:business_id/messages", handlers.GetCustomerMessages)
		customerProtected.POST("/businesses/:business_id/messages", handlers.CreateCustomerMessage)
		customerProtected.POST("/heartbeat", handlers.CustomerHeartbeat)
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
