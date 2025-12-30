package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shjung-dev/ChatApplication/backend/config"
	"github.com/shjung-dev/ChatApplication/backend/helpers"
	"github.com/shjung-dev/ChatApplication/backend/models"
	"go.mongodb.org/mongo-driver/bson"
)

func getClaim(token string) (*helpers.Claims, bool) {
	claims, err := helpers.ValidateToken(token)
	if err != nil {
		return nil, false
	}
	return claims, true
}

func Request() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		token := c.Query("token")

		claims, ok := getClaim(token)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		receiver := c.Param("receiver")
		sender := claims.Username
		request := models.Request{
			From:   sender,
			To:     receiver,
			Status: "pending",
		}
		requestCollection := config.OpenCollection("request")

		_, insertErr := requestCollection.InsertOne(ctx, request)

		if insertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": insertErr.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Requested",
		})
	}
}

func Reject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		token := c.Query("token")

		claims, ok := getClaim(token)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		person_rejecting := claims.Username
		person_getting_rejected := c.Param("receiver")
		
		filter := bson.M{
			"from": person_getting_rejected,
			"to":   person_rejecting,
		}

		requestCollection := config.OpenCollection("request")

		result, err := requestCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No document found to be deleted"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Rejected",
		})
	}
}

func Accept() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		token := c.Query("token")

		claims, ok := getClaim(token)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		accepter := claims.Username
		sender := c.Param("username")

		friend := models.Friend{
			Username:       &accepter,
			FriendUsername: &sender,
		}

		friendCollection := config.OpenCollection("friend")
		_, insertErr := friendCollection.InsertOne(ctx, friend)

		if insertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": insertErr.Error()})
			return
		}

		filter := bson.M{
			"from": sender,
			"to":   accepter,
		}

		update := bson.M{
			"$set": bson.M{
				"status": "accepted",
			},
		}

		requestCollection := config.OpenCollection("request")

		result, err := requestCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Accepted"})
	}
}


func Remove() gin.HandlerFunc{
	return func(c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		token := c.Query("token")

		claims, ok := getClaim(token)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		person_removing := claims.Username
		person_getting_removed := c.Param("username")

		filter := bson.M{
			"username" : person_removing,
			"friendusername": person_getting_removed,
		}

		friendCollection := config.OpenCollection("friend")

		result, err := friendCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No document found to be deleted"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Removed",
		})
	}
}