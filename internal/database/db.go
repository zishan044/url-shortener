package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect(dsn string) {
	var err error
	ctx := context.Background()

	Pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Couldn't connect to postgres: %v\n", err)
	}

	err = Pool.Ping(ctx); if err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	log.Println("Connected to postgres via pgxpool!")
}