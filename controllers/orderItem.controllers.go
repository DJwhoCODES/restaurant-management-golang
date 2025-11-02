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

var orderItemModel *mongo.Collection = database.OpenCollection(database.MongoClient, "order_item")

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pageStr := c.Query("page")
		limitStr := c.Query("limit")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}

		skip := int64((page - 1) * limit)
		limit64 := int64(limit)

		opts := options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(skip).
			SetLimit(limit64)

		cursor, err := orderItemModel.Find(ctx, bson.M{}, opts)
		if err != nil {
			log.Printf("Error fetching order items: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order items"})
			return
		}
		defer cursor.Close(ctx)

		var orderItems []bson.M
		if err := cursor.All(ctx, &orderItems); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding order items"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"page":  page,
			"limit": limit,
			"data":  orderItems,
		})
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		orderItemId := c.Param("order_item_id")
		var orderItem models.OrderItem

		err := orderItemModel.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order item"})
			return
		}

		c.JSON(http.StatusOK, orderItem)
	}
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var orderItem models.OrderItem
		if err := c.ShouldBindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		orderItem.ID = primitive.NewObjectID()
		orderItem.Order_Item_Id = orderItem.ID.Hex()
		orderItem.Created_At = time.Now()
		orderItem.Updated_At = time.Now()

		_, err := orderItemModel.InsertOne(ctx, orderItem)
		if err != nil {
			log.Printf("Error inserting order item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating order item"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Order item created successfully",
			"data":    orderItem,
		})
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		orderItemId := c.Param("order_item_id")

		var orderItem models.OrderItem
		if err := c.ShouldBindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		updateObj := bson.D{}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{Key: "quantity", Value: *orderItem.Quantity})
		}
		if orderItem.Unit_Price != nil {
			updateObj = append(updateObj, bson.E{Key: "unit_price", Value: *orderItem.Unit_Price})
		}
		if orderItem.Food_Id != nil {
			updateObj = append(updateObj, bson.E{Key: "food_id", Value: *orderItem.Food_Id})
		}
		if orderItem.Order_Id != "" {
			updateObj = append(updateObj, bson.E{Key: "order_id", Value: orderItem.Order_Id})
		}

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: time.Now()})

		result, err := orderItemModel.UpdateOne(
			ctx,
			bson.M{"order_item_id": orderItemId},
			bson.D{{Key: "$set", Value: updateObj}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating order item"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order item updated successfully"})
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		orderId := c.Param("order_id")
		cursor, err := orderItemModel.Find(ctx, bson.M{"order_id": orderId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching items for order"})
			return
		}
		defer cursor.Close(ctx)

		var items []bson.M
		if err := cursor.All(ctx, &items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding order items"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"order_id": orderId,
			"items":    items,
		})
	}
}

func ItemsByOrder(orderID string) (orderItems []bson.M, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "order_id", Value: orderID}}}},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "food"},
				{Key: "localField", Value: "food_id"},
				{Key: "foreignField", Value: "food_id"},
				{Key: "as", Value: "food_details"},
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$food_details"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
	}

	cursor, err := orderItemModel.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &orderItems); err != nil {
		return nil, err
	}

	return orderItems, nil
}
