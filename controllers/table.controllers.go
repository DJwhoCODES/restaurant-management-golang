package controllers

import (
	"github.com/djwhocodes/restaurant_management/database"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var tableModel *mongo.Collection = database.OpenCollection(database.MongoClient, "table")

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {}
}
