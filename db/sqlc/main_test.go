package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)
var testQueries *Queries
var testDB *pgx.Conn

func TestMain(m *testing.M) {
    conn, err := pgx.Connect(context.Background(), "postgresql://postgres:password@localhost:5432/tradingdb?sslmode=disable")
    if err != nil {
        log.Fatal("cannot connect to db:", err)
    }
    testDB = conn
    testQueries = New(conn)


    os.Exit(m.Run())
}