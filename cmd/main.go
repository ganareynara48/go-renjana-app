package main

import (
	"renjana-app/config"
	"renjana-app/database"
	"renjana-app/routes"

	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	database.ConnectDB()

	router := gin.Default()

	router.Static("/storage/upload", "./storage/upload")

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hola, welcome to REST API Renjana Coffee and Roastery"})
	})

	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	routes.MasterProductRoutes(router)
	routes.ProductCategoryRoutes(router)
	routes.BasketRoutes(router)

	router.Run("127.0.0.1:9090")
}
