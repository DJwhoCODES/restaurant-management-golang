package helpers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/djwhocodes/restaurant_management/database"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Uid       string `json:"uid"`
	jwt.RegisteredClaims
}

var userModel *mongo.Collection = database.OpenCollection(database.MongoClient, "user")
var SECRET_KEY = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email, firstName, lastName, uid string) (accessToken, refreshToken string, err error) {
	accessClaims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Uid:       uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	accessToken, err = accessTokenObj.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err = refreshTokenObj.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func UpdateAllTokens(accessToken, refreshToken, userId string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateObj := bson.D{
		{Key: "token", Value: accessToken},
		{Key: "refresh_token", Value: refreshToken},
		{Key: "updated_at", Value: time.Now()},
	}

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userModel.UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: updateObj}},
		&opt,
	)

	if err != nil {
		fmt.Println("Error updating tokens:", err)
		return
	}
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "token is invalid"
		return
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		msg = "token has expired"
		return
	}

	return claims, msg
}
