package routes

import (
	"github.com/djwhocodes/restaurant_management/controllers"
	"github.com/gin-gonic/gin"
)

func FoodRoutes(router *gin.Engine) {
	router.GET("/foods", controllers.GetFoods())
	router.GET("/foods/:food_id", controllers.GetFood())
	router.POST("/foods", controllers.CreateFood())
	router.PATCH("/foods/:food_id", controllers.UpdateFood())
}
