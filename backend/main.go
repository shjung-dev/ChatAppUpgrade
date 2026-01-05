package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shjung-dev/ChatApplication/backend/config"
	"github.com/shjung-dev/ChatApplication/backend/helpers"
	"github.com/shjung-dev/ChatApplication/backend/network"
	"github.com/shjung-dev/ChatApplication/backend/routes"
)

func main() {

	/*
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	*/

	port := os.Getenv("PORT")
	jwtKey := os.Getenv("JWT_KEY")
	uri := os.Getenv("MONGO_URI")

	config.ConnectDatabase(uri)

	helpers.SetJWTKey(jwtKey)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/ws", func(c *gin.Context) {
		token := c.Query("token")

		if token ==""{
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		
		claims , err := helpers.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized , gin.H{"error":"Invalid token"})
			return
		}

		username := claims.Username

		roomID:=network.GenerateRoomName(username)
		personalRoom := network.GetRoom(roomID)

		req := c.Request

		q := req.URL.Query()
		q.Set("username", username)
		req.URL.RawQuery = q.Encode()
		
		personalRoom.ServeHttp(c.Writer , req)
	})

	routes.SetUpRoutes(r)

	log.Println("Server is running on localhost:" + port)
	r.Run(":" + port)

}
