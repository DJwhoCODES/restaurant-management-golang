package routes

import (
	"github.com/djwhocodes/restaurant_management/controllers"
	"github.com/gin-gonic/gin"
)

func OrderItemRoutes(router *gin.Engine) {
	router.GET("/orderItems", controllers.GetOrderItems())
	router.GET("/orderItems/:orderItem_id", controllers.GetOrderItem())
	router.POST("/orderItems", controllers.CreateOrderItem())
	router.PATCH("/orderItems/:orderItem_id", controllers.UpdateOrderItem())
	router.GET("/orderItems-order/:order_id", controllers.GetOrderItemsByOrder())
}
