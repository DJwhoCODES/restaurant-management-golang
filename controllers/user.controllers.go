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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var userModel *mongo.Collection = database.OpenCollection(database.MongoClient, "user")

func GetUsers() gin.HandlerFunc {
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
		if err != nil || limit < 1 {
			limit = 10
		}

		skip := (page - 1) * limit

		opts := options.Find().
			SetSkip(int64(skip)).
			SetLimit(int64(limit)).
			SetSort(bson.D{{Key: "created_at", Value: -1}})

		cursor, err := userModel.Find(ctx, bson.M{}, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
			return
		}
		defer cursor.Close(ctx)

		var users []bson.M
		if err := cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding user data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"page":  page,
			"limit": limit,
			"count": len(users),
			"data":  users,
		})
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		err := userModel.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		//convert the json data incoming from postman to something golang understands
		//validate the data based on the defined struct
		//check if email is already registered
		//hash password
		//check if phone exists
		//get some extra details - created_at ...
		//generate tokens
		//insert into db
		//return statusOK
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		//convert the login data in go understandable format
		//find a user with that email
		//verify the password
		//generate tokens
		//update tokens
		//return statusOK
	}
}

// func HashPassword(password string) string {

// }

// func VerifyPassword(actualPassword, providedPassword string) (bool, string) {

// }
