package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/djwhocodes/restaurant_management/database"
	"github.com/djwhocodes/restaurant_management/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var orderModel *mongo.Collection = database.OpenCollection(database.MongoClient, "order")

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}
		skip := (page - 1) * limit

		findOptions := options.Find()
		findOptions.SetSkip(int64(skip))
		findOptions.SetLimit(int64(limit))

		cursor, err := orderModel.Find(ctx, bson.M{}, findOptions)
		if err != nil {
			log.Printf("Error fetching orders: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching orders"})
			return
		}
		defer cursor.Close(ctx)

		var orders []bson.M
		if err := cursor.All(ctx, &orders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding orders"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"page":   page,
			"limit":  limit,
			"orders": orders,
		})
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		orderId := c.Param("order_id")
		if orderId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order_id parameter is required"})
			return
		}

		var order models.Order
		err := orderModel.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			} else {
				log.Printf("Error fetching order (id=%s): %v", orderId, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order"})
			}
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var order models.Order
		var table models.Table

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if order.Table_Id != nil {
			err := tableModel.FindOne(ctx, bson.M{"table_id": *order.Table_Id}).Decode(&table)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
				return
			}
		}

		order.Created_At = time.Now()
		order.Updated_At = order.Created_At
		order.Order_Date = order.Created_At
		order.ID = primitive.NewObjectID()
		order.Order_Id = order.ID.Hex()

		result, err := orderModel.InsertOne(ctx, order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while creating order"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Order created successfully",
			"data":    result,
		})
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		orderId := c.Param("order_id")
		if orderId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order_id parameter is required"})
			return
		}

		var order models.Order
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		if order.Table_Id != nil {
			var table models.Table
			err := tableModel.FindOne(ctx, bson.M{"table_id": *order.Table_Id}).Decode(&table)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
				return
			}
		}

		updateObj := bson.D{}

		if order.Table_Id != nil {
			updateObj = append(updateObj, bson.E{Key: "table_id", Value: *order.Table_Id})
		}
		if !order.Order_Date.IsZero() {
			updateObj = append(updateObj, bson.E{Key: "order_date", Value: order.Order_Date})
		}

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: time.Now().UTC()})

		if len(updateObj) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields provided for update"})
			return
		}

		filter := bson.M{"order_id": orderId}
		update := bson.D{{Key: "$set", Value: updateObj}}

		result, err := orderModel.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Error updating order (id=%s): %v", orderId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Order updated successfully",
		})
	}
}
