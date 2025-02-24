package main

import (
	"log"
	"os"

	"github.com/buranasakS/trading_application/config"
	_ "github.com/buranasakS/trading_application/docs"
	"github.com/buranasakS/trading_application/routes"
	"github.com/gin-gonic/gin"
)

// @title Trading Application API
// @version 1.0
// @description This is a Golang Application For Backend Candidate Test

// @host localhost:8080
// @BasePath /
func main() {

	db := config.ConnectDatabase()
	defer config.CloseDatabase(db)

	router := gin.Default()
	routes.SetupRoutes(router)

	port := os.Getenv("PORT")
	err := router.Run(":" + port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
