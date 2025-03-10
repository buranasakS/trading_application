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
    conn, err := pgx.Connect(context.Background(), "postgresql://postgres:password@localhost:5432/tradingdb_test?sslmode=disable")
    if err != nil {
        log.Fatal("cannot connect to db:", err)
    }
    defer conn.Close(context.Background())

    testDB = conn
    testQueries = New(conn)

    clearTestDB()

    os.Exit(m.Run())
}

func clearTestDB() {
    _, err := testDB.Exec(context.Background(), `
        TRUNCATE TABLE users, products, commissions, affiliates RESTART IDENTITY CASCADE;
    `)
    if err != nil {
        log.Fatal("failed to clear test db:", err)
    }
}