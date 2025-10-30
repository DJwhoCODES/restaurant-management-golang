package routes

import (
	"github.com/djwhocodes/restaurant_management/controllers"
	"github.com/gin-gonic/gin"
)

func TableRoutes(router *gin.Engine) {
	router.GET("/tables", controllers.GetTables())
	router.GET("/tables/:table_id", controllers.GetTable())
	router.POST("/tables", controllers.CreateTable())
	router.PATCH("/tables/:table_id", controllers.UpdateTable())
}
