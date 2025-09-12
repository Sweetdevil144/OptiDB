package db

import (
	"context"
	"fmt"
	"internal/config"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectOrFail() *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), config.Config("DB_URI"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	return conn
}
