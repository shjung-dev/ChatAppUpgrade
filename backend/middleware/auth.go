package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shjung-dev/ChatApplication/backend/config"
	"github.com/shjung-dev/ChatApplication/backend/helpers"
	"github.com/shjung-dev/ChatApplication/backend/models"
	"go.mongodb.org/mongo-driver/bson"
)

func Authenticate() gin.HandlerFunc {

	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		//Header usually in the form of -> Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		//Remove Bearer because we are only interested in the token
		authHeader = strings.TrimPrefix(authHeader, "Bearer ") //Remove the space in front of Bearer so that we get the pure <token> part

		claims, err := helpers.ValidateToken(authHeader)

		if err != nil {
			if errors.Is(err, helpers.ErrTokenExpired) {
				//Token expired - client should call /refresh
				c.JSON(401, gin.H{"error": "token expired"})
				c.Abort()
				return
			}
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		//Check if the token matches the token in the database
		userID := claims.UserID
		var user models.User

		userCollection := config.OpenCollection("user")
		err = userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user is not found"})
			c.Abort()
			return
		}

		if *user.Token != authHeader {
			c.JSON(401, gin.H{"error": "Wrong access token used"})
			c.Abort()
			return
		}

		//Token valid
		c.Set("claims", claims)
		c.Next()

	}
}
