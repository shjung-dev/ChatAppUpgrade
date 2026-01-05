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

func Reject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		claims, ok := c.Get("claims")

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		person_rejecting := claims.(*helpers.Claims).Username
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

		claims, ok := c.Get("claims")

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		accepter := claims.(*helpers.Claims).Username
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

func Remove() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		claims, ok := c.Get("claims")

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		person_removing := claims.(*helpers.Claims).Username
		person_getting_removed := c.Param("username")

		filterOne := bson.M{
			"username":       person_removing,
			"friendusername": person_getting_removed,
		}

		filterTwo := bson.M{
			"from": person_getting_removed,
			"to":   person_removing,
		}

		friendCollection := config.OpenCollection("friend")

		result, err := friendCollection.DeleteOne(ctx, filterOne)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No friend document found to be deleted"})
			return
		}

		requestCollection := config.OpenCollection("request")

		result, e := requestCollection.DeleteOne(ctx, filterTwo)
		if e != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": e})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No request document found to be deleted"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Friend removed and Request Record Deleted",
		})
	}
}
