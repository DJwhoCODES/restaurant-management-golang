package routes

import (
	"github.com/djwhocodes/restaurant_management/controllers"
	"github.com/gin-gonic/gin"
)

func NoteRoutes(router *gin.Engine) {
	router.GET("/notes", controllers.GetNotes())
	router.GET("/notes/:note_id", controllers.GetNote())
	router.POST("/notes", controllers.CreateNote())
	router.PATCH("/notes/:note_id", controllers.UpdateNote())
}
