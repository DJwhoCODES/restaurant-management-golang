package routes

import (
	"github.com/djwhocodes/restaurant_management/controllers"
	"github.com/gin-gonic/gin"
)

func MenuRoutes(router *gin.Engine) {
	router.GET("/menu", controllers.GetMenus())
	router.GET("/menu/:menu_id", controllers.GetMenu())
	router.POST("/menu", controllers.CreateMenu())
	router.PATCH("/menu/:menu_id", controllers.UpdateMenu())
}
