package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := menuModel.Find(ctx, bson.M{})
		if err != nil {
			log.Printf("Error fetching menus: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch menus"})
			return
		}
		defer cursor.Close(ctx)

		var menus []bson.M
		if err = cursor.All(ctx, &menus); err != nil {
			log.Printf("Error decoding menu documents: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode menus"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Menus retrieved successfully",
			"data":    menus,
		})
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {}
}
