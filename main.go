package main

import (
	"fmt"
	"log"
	"os"

	"github.com/djwhocodes/restaurant_management/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// go get github.com/joho/godotenv

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or couldn't load it")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	// router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.InvoiceRoutes(router)
	routes.MenuRoutes(router)
	routes.NoteRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.TableRoutes(router)
	routes.UserRoutes(router)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, Go!",
		})
	})

	fmt.Println("Server running on port:", port)
	router.Run(":" + port)
}
