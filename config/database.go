package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type PgDatabase struct {
	DB *pgx.Conn
}

func ConnectDatabase() *PgDatabase {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	db, err := pgx.Connect(ctx, "postgresql://postgres:password@localhost:5432/tradingdb?sslmode=disable")
	if err != nil {
		log.Fatalf("Cannot connect to database: %v\n", err)
	}

	ctx, cancle = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	err = db.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	return &PgDatabase{DB: db}
}

func CloseDatabase(db *PgDatabase) {
	if db.DB != nil {
		err := db.DB.Close(context.Background())
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			fmt.Println("Database connection closed")
		}
	}
}
