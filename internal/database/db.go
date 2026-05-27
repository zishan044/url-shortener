package database

import (
	"context"
	"embed"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*
var migrationFiles embed.FS

var Pool *pgxpool.Pool

func Connect(dsn string) {
	var err error
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		log.Printf("Connecting to database (attempt %d/5)...", i)
		Pool, err = pgxpool.New(ctx, dsn)
		
		if err == nil {
			err = Pool.Ping(ctx)
			if err == nil {
				log.Println("Connected to postgres via pgxpool!")
				break
			}
		}

		log.Printf("Database connection failed, retrying in 2 seconds... Error: %v", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database after 5 attempts: %v\n", err)
	}

	RunMigrations(dsn)
}

func RunMigrations(dsn string) {
	d, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		log.Fatalf("migration init failed: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		log.Fatalf("migrate init failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migrations applied")
}