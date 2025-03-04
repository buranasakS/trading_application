package main

import (
	"log"
	"os"
	"testing"

	"github.com/buranasakS/trading_application/config"
	"github.com/buranasakS/trading_application/routes"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
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