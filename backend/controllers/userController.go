package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shjung-dev/ChatApplication/backend/config"
	"github.com/shjung-dev/ChatApplication/backend/helpers"
	"github.com/shjung-dev/ChatApplication/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var validate = validator.New()

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if validationErr := validate.Struct(user); validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		userCollection := config.OpenCollection("user")
		count, err := userCollection.CountDocuments(ctx, bson.M{"username": user.Username})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
			return
		}

		user.Password = helpers.HashPassword(user.Password)
		user.Created_at = time.Now()
		user.Updated_at = time.Now()
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		accessToken, refreshToken := helpers.GenerateToken(user.User_id, *user.Username)
		user.Token = &accessToken
		user.Refresh_token = &refreshToken

		_, insertErr := userCollection.InsertOne(ctx, user)

		if insertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": insertErr.Error()})
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "signup successful",
			"user":          user,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userCollection := config.OpenCollection("user")
		err := userCollection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username. User is not found"})
			return
		}

		passwordIsValid, msg := helpers.VerifyPassword(*foundUser.Password, *user.Password)

		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		}

		token, refreshToken := helpers.GenerateToken(foundUser.User_id, *foundUser.Username)

		helpers.UpdateAllToken(token, refreshToken, foundUser.User_id)
		
		c.JSON(http.StatusOK, gin.H{
			"message":       "login successful",
			"user":          foundUser,
			"access_token":  token,
			"refresh_token": refreshToken,
		})
	}
}

func SearchUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		claims, ok := c.Get("claims")

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		sender := claims.(*helpers.Claims).Username

		receiver := c.Param("receiver")

		var user models.User

		userCollection := config.OpenCollection("user")

		err := userCollection.FindOne(ctx, bson.M{"username": receiver}).Decode(&user)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var request models.Request
       
		filter := bson.M{
			"from": sender,
			"to":   receiver,
		}

		requestCollection := config.OpenCollection("request")

		err = requestCollection.FindOne(ctx, filter).Decode(&request)

		if err != nil {
			//Request has not been sent to the receiver yet
			fmt.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"message":  "available",
				"receiver": user,
			})
			return
		}

		//Request has already been sent and it is not accepted yet
		if request.Status == "pending" {
			c.JSON(http.StatusOK, gin.H{
				"message":  "pending",
				"receiver": user,
			})
			return
		}

		//Request has already been sent and it is accepted already
		if request.Status == "accepted" {
			c.JSON(http.StatusOK, gin.H{
				"message":  "accepted",
				"receiver": user,
			})
			return
		}
	}
}

func RefreshTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		//Sends request with refresh_token as Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token is required"})
			c.Abort()
			return
		}

		authHeader = strings.TrimPrefix(authHeader, "Bearer ")

		//Validate Token
		claims, err := helpers.ValidateToken(authHeader)
		if err != nil {
			//Refresh token is invalid or expired -> force user to login back
			c.JSON(http.StatusUnauthorized, gin.H{"error": "relogin"})
			return
		}

		userID := claims.UserID

		var user models.User

		userCollection := config.OpenCollection("user")
		err = userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)

		if err != nil {
			//User not found
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user is not found"})
			return
		}

		//Check if the refresh_token matches the one in the database as well
		if *user.Refresh_token != authHeader {
			//Wrong refresh_token is used
			c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong refresh token is used"})
			return
		}

		//Generate new tokens
		newAccessToken, newRefreshToken := helpers.GenerateToken(user.User_id, *user.Username)

		helpers.UpdateAllToken(newAccessToken, newRefreshToken, user.User_id)
		
		c.JSON(http.StatusOK, gin.H{
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken,
		})
	}
}
