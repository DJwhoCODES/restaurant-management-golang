package controllers

import (
	"context"
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

var tableModel *mongo.Collection = database.OpenCollection(database.MongoClient, "table")

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")

		page, _ := strconv.ParseInt(pageStr, 10, 64)
		limit, _ := strconv.ParseInt(limitStr, 10, 64)

		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}

		skip := (page - 1) * limit

		opts := options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(skip).
			SetLimit(limit)

		cursor, err := tableModel.Find(ctx, bson.M{}, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching tables"})
			return
		}
		defer cursor.Close(ctx)

		var tables []bson.M
		if err = cursor.All(ctx, &tables); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding tables"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  tables,
			"page":  page,
			"limit": limit,
		})
	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		tableId := c.Param("table_id")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var table models.Table
		err := tableModel.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}

		c.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		validate := validator.New()

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := validate.Struct(table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		table.ID = primitive.NewObjectID()
		table.Created_At = time.Now()
		table.Updated_At = time.Now()
		table.Table_Id = table.ID.Hex()

		_, err := tableModel.InsertOne(ctx, table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating table"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Table created successfully",
			"data":    table,
		})
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		tableId := c.Param("table_id")
		var table models.Table

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		updateData := bson.M{
			"updated_at": time.Now(),
		}

		if table.Number_Of_Guests != nil {
			updateData["number_of_guests"] = table.Number_Of_Guests
		}
		if table.Table_Number != nil {
			updateData["table_number"] = table.Table_Number
		}

		result, err := tableModel.UpdateOne(
			ctx,
			bson.M{"table_id": tableId},
			bson.M{"$set": updateData},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating table"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Table updated successfully"})
	}
}
