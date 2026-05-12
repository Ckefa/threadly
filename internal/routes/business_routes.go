package routes

import (
	"threadly/internal/db"
	"threadly/internal/handlers"
	"threadly/internal/handlers/business"
	"threadly/internal/handlers/client"
	"threadly/internal/middleware"

	"github.com/gin-gonic/gin"
)

var businessHandler *business.BusinessHandler

func SetupBusinessRoutes(r *gin.Engine) {

	// Initialize business handler
	businessHandler = business.NewBusinessHandler(db.DB)

	// PUBLIC - Business Auth Routes
	r.GET("/business/login", handlers.ShowLogin)
	r.GET("/business/register", handlers.ShowRegister)
	r.POST("/business/login", handlers.Login)
	r.POST("/business/register", handlers.Register)
	r.GET("/business/logout", handlers.Logout)

	// PROTECTED BUSINESS ROUTES
	protected := r.Group("/business")
	protected.Use(middleware.BizzMiddleware())
	{
		// Business Dashboard routes
		protected.GET("/", businessHandler.GetBizHome)
		protected.GET("/dashboard", businessHandler.GetDashboard)
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
		protected.POST("/clients", client.CreateClient)
		protected.DELETE("/clients/:id", client.DeleteClient)
		protected.PUT("/clients/:id/status", client.UpdateClientStatus)
		protected.GET("/clients/:id/conversation-id", client.GetClientConversationID)

		// Business Message routes
		protected.GET("/clients/:id/messages", handlers.GetMessages)
		protected.POST("/clients/:id/messages", handlers.CreateMessage)
		protected.PUT("/messages/:message_id", handlers.UpdateMessage)
		protected.PUT("/clients/:id/read", handlers.MarkConversationAsRead)

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
		protected.POST("/clients/:id/quick-booking", business.QuickBooking)
		protected.POST("/clients/:id/quick-order", business.QuickOrder)
		protected.POST("/clients/:id/request-payment", business.RequestPayment)
		protected.POST("/clients/:id/set-goal", business.SetGoal)

		// Product picker & order lifecycle routes
		protected.GET("/conversations/:conversation_id/products", businessHandler.GetConversationProducts)
		protected.GET("/conversations/:conversation_id/services", businessHandler.GetConversationServices)
		protected.POST("/conversations/:conversation_id/order-draft", businessHandler.CreateOrderDraft)
		protected.POST("/orders/:id/send", businessHandler.SendOrderToClient)
		protected.POST("/orders/:id/confirm", businessHandler.ConfirmOrderBusiness)
		protected.POST("/orders/:id/reject", businessHandler.RejectOrder)

		// Conversation progress routes
		protected.GET("/conversations/:conversation_id/progress", handlers.GetConversationProgress)
		protected.PUT("/conversations/:conversation_id/stage", handlers.UpdateConversationStage)
	}

}
