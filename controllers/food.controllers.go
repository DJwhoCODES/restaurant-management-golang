package controllers

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/djwhocodes/restaurant_management/database"
	"github.com/djwhocodes/restaurant_management/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodModel *mongo.Collection = database.OpenCollection(database.MongoClient, "food")
var menuModel *mongo.Collection = database.OpenCollection(database.MongoClient, "menu")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		foodID := c.Param("food_id")
		if foodID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "food_id parameter is required"})
			return
		}

		var food models.Food
		err := foodModel.FindOne(ctx, bson.M{"food_id": foodID}).Decode(&food)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.JSON(http.StatusNotFound, gin.H{"error": "food item not found"})
			} else {
				log.Printf("Error fetching food (id=%s): %v", foodID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch food item"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "food item retrieved successfully",
			"data":    food,
		})
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var food models.Food

		if err := c.ShouldBindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
			return
		}

		if err := validate.Struct(food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		var menu models.Menu
		if err := menuModel.FindOne(ctx, bson.M{"menu_id": food.Menu_Id}).Decode(&menu); err != nil {
			log.Printf("Menu not found for ID: %v, error: %v", food.Menu_Id, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		}

		now := time.Now().UTC()
		food.ID = primitive.NewObjectID()
		food.Food_Id = food.ID.Hex()
		food.Created_At = now
		food.Updated_At = now

		if food.Price != nil {
			num := toFixed(*food.Price, 2)
			food.Price = &num
		}

		result, err := foodModel.InsertOne(ctx, food)
		if err != nil {
			log.Printf("Error inserting food item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create food item"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "Food item created successfully",
			"data":    result,
		})
	}
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

// func round(num float64) int {}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(num*output) / output
}
