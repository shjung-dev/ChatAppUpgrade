package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shjung-dev/ChatApplication/backend/config"
	"github.com/shjung-dev/ChatApplication/backend/models"
)

func CreateRoom() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		token := c.Query("token")

		claims, ok := getClaim(token)

		hostUserName := claims.Username

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		var body struct {
			Members []string `json:"members"`
		}

		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		groupID := uuid.New().String()

		body.Members = append(body.Members, hostUserName)

		group := models.Group{
			Hostname: &hostUserName,
			GroupID: &groupID,
			Members: body.Members,
		}
		
		roomCollection := config.OpenCollection("group")

		_, insertErr := roomCollection.InsertOne(ctx, group)

		if insertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": insertErr.Error()})
		}

		

		c.JSON(http.StatusOK , gin.H{
			"message":"Room Created",
			"groupID" : groupID,
		})
	}
}
