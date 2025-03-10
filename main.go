package main

import (
	"log"
	"os"

	"github.com/buranasakS/trading_application/config"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/buranasakS/trading_application/handlers"
	"github.com/buranasakS/trading_application/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/buranasakS/trading_application/docs" 

)

// @title Trading Application API
// @version 1.0
// @description This is a Golang Application For Backend Candidate Test
// @host localhost:8080
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /
func main() {

	database := config.ConnectDatabase()
	defer config.CloseDatabase(database)

	queries := db.New(database.DB)	
	h := handlers.NewHandler(queries)

	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.Use(gin.Recovery())

	routes.SetupRoutes(router, h)
	
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	port := os.Getenv("PORT")
	err = router.Run(":" + port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
