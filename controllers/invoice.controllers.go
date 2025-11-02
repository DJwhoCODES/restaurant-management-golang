package controllers

import (
	"context"
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

var invoiceModel *mongo.Collection = database.OpenCollection(database.MongoClient, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}
		skip := (page - 1) * limit

		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
		findOptions.SetSkip(int64(skip))
		findOptions.SetLimit(int64(limit))

		cursor, err := invoiceModel.Find(ctx, bson.M{}, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching invoices"})
			return
		}
		defer cursor.Close(ctx)

		var invoices []bson.M
		if err := cursor.All(ctx, &invoices); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding invoices"})
			return
		}

		total, _ := invoiceModel.CountDocuments(ctx, bson.M{})

		c.JSON(http.StatusOK, gin.H{
			"data":       invoices,
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		})
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		invoiceId := c.Param("invoice_id")

		var invoice bson.M
		err := invoiceModel.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": invoice})
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var invoice models.Invoice
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		var order models.Order
		err := orderModel.FindOne(ctx, bson.M{"order_id": invoice.Order_Id}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_Id = invoice.ID.Hex()
		invoice.Created_At = time.Now()
		invoice.Updated_At = invoice.Created_At

		if invoice.Payment_Status == nil {
			status := "PENDING"
			invoice.Payment_Status = &status
		}

		result, err := invoiceModel.InsertOne(ctx, invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating invoice"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Invoice created successfully",
			"data":    result,
		})
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		invoiceId := c.Param("invoice_id")
		var invoice models.Invoice
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		updateFields := bson.D{}
		if invoice.Payment_Status != nil {
			updateFields = append(updateFields, bson.E{Key: "payment_status", Value: invoice.Payment_Status})
		}
		if invoice.Payment_Method != nil {
			updateFields = append(updateFields, bson.E{Key: "payment_method", Value: invoice.Payment_Method})
		}
		updateFields = append(updateFields, bson.E{Key: "updated_at", Value: time.Now()})

		filter := bson.M{"invoice_id": invoiceId}
		update := bson.D{{Key: "$set", Value: updateFields}}

		result, err := invoiceModel.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating invoice"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Invoice updated successfully"})
	}
}
