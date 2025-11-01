package controllers

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/djwhocodes/restaurant_management/database"
	"github.com/djwhocodes/restaurant_management/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodModel *mongo.Collection = database.OpenCollection(database.MongoClient, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			limit = 10
		}

		skip := (page - 1) * limit

		findOptions := options.Find()
		findOptions.SetSkip(int64(skip))
		findOptions.SetLimit(int64(limit))
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

		cursor, err := foodModel.Find(ctx, bson.M{}, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching food items"})
			return
		}
		defer cursor.Close(ctx)

		var foods []bson.M
		if err := cursor.All(ctx, &foods); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding food items"})
			return
		}

		total, err := foodModel.CountDocuments(ctx, bson.M{})
		if err != nil {
			total = int64(len(foods))
		}

		c.JSON(http.StatusOK, gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
			"data":       foods,
		})
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
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		foodId := c.Param("food_id")
		if foodId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "food_id parameter is required"})
			return
		}

		var food models.Food
		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if food.Menu_Id != nil {
			var menu models.Menu
			err := menuModel.FindOne(ctx, bson.M{"menu_id": *food.Menu_Id}).Decode(&menu)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
				return
			}

			if time.Now().Before(menu.Start_Date) || time.Now().After(menu.End_Date) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Menu is not active"})
				return
			}
		}

		updateObj := bson.D{}

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: *food.Name})
		}
		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: *food.Price})
		}
		if food.Food_Image != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: *food.Food_Image})
		}
		if food.Menu_Id != nil {
			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: *food.Menu_Id})
		}

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: time.Now()})

		if len(updateObj) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields provided for update"})
			return
		}

		filter := bson.M{"food_id": foodId}
		update := bson.D{{Key: "$set", Value: updateObj}}

		result, err := foodModel.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating food item"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Food item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Food item updated successfully"})
	}
}

// func round(num float64) int {}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(num*output) / output
}
