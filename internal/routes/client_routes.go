package routes

import (
	"threadly/internal/handlers"
	"threadly/internal/handlers/client"

	"github.com/gin-gonic/gin"
)

func SetupClientRoutes(r *gin.Engine) {

	// PUBLIC - client Route
	r.GET("/client/login", client.ShowClientLogin)
	r.POST("/client/send-otp", client.SendClientOTP)
	r.POST("/client/verify-otp", client.VerifyClientOTP)
	r.GET("/client/logout", client.ClientLogout)

	// PROTECTED client ROUTES
	clientProtected := r.Group("/client")
	clientProtected.Use(client.ClientMiddleware())
	{
		clientProtected.GET("/", client.ClientDashboard)
		clientProtected.GET("/businesses/:business_id/messages", client.GetClientMessages)
		clientProtected.POST("/businesses/:business_id/messages", client.CreateClientMessage)
		clientProtected.GET("/businesses/:business_id/products", businessHandler.GetBusinessProducts)
		clientProtected.GET("/businesses/:business_id/services", businessHandler.GetBusinessServices)
		clientProtected.POST("/businesses/:business_id/bookings", businessHandler.ClientCreateBooking)
		clientProtected.POST("/orders", businessHandler.ClientCreateOrder)
		clientProtected.POST("/bookings", businessHandler.ClientCreateBooking)
		clientProtected.POST("/orders/:id/update", client.ClientUpdateOrder)
		clientProtected.POST("/bookings/:id/update", client.ClientUpdateBooking)
		clientProtected.PUT("/businesses/:business_id/read", handlers.MarkClientConversationAsRead)
		clientProtected.POST("/heartbeat", client.ClientHeartbeat)
	}

}
