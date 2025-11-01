package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/djwhocodes/restaurant_management/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		menuID := c.Param("menu_id")
		if menuID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "menu_id parameter is required"})
			return
		}

		var menu models.Menu
		err := menuModel.FindOne(ctx, bson.M{"menu_id": menuID}).Decode(&menu)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
			} else {
				log.Printf("Error fetching menu (id=%s): %v", menuID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menu"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "menu retrieved successfully",
			"data":    menu,
		})
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var menu models.Menu

		if err := c.ShouldBindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON payload",
				"details": err.Error(),
			})
			return
		}

		if err := validate.Struct(menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": err.Error(),
			})
			return
		}

		now := time.Now().UTC()
		menu.ID = primitive.NewObjectID()
		menu.Menu_Id = menu.ID.Hex()
		menu.Created_At = now
		menu.Updated_At = now

		result, err := menuModel.InsertOne(ctx, menu)
		if err != nil {
			log.Printf("Error inserting menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create menu",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "Menu created successfully",
			"data":    result,
		})
	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		menuID := c.Param("menu_id")
		if menuID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "menu_id parameter is required"})
			return
		}

		var menu models.Menu
		if err := c.ShouldBindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON payload",
				"details": err.Error(),
			})
			return
		}

		if !menu.Start_Date.IsZero() && !menu.End_Date.IsZero() {
			if !inTimeSpan(menu.Start_Date, menu.End_Date, time.Now()) {
				msg := "Invalid time range: current time not in between start and end date"
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
		}

		var updateObj primitive.D

		if menu.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
		}
		if menu.Category != nil {
			updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
		}
		if !menu.Start_Date.IsZero() {
			updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_Date})
		}
		if !menu.End_Date.IsZero() {
			updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_Date})
		}

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: time.Now().UTC()})

		if len(updateObj) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields provided for update"})
			return
		}

		result, err := menuModel.UpdateOne(
			ctx,
			bson.M{"menu_id": menuID},
			bson.D{{Key: "$set", Value: updateObj}},
		)
		if err != nil {
			log.Printf("Error updating menu (id=%s): %v", menuID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Menu updated successfully",
		})
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return (check.Equal(start) || check.After(start)) && (check.Equal(end) || check.Before(end))
}
