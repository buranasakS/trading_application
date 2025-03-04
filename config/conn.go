package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)


type Database struct {
	DB *pgx.Conn
}

func ConnectDatabase() *Database {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DB_SOURCE")
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Cannot connect to database: %v\n", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	return &Database{DB: conn}
}

func CloseDatabase(db *Database) {
	if db.DB != nil {
		err := db.DB.Close(context.Background())
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			fmt.Println("Database connection closed")
		}
	}
}
