package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"

	"covid19-greece-api/pkg/env"
)

func InitPostgresDb(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := env.EnvOrDefault(
		"POSTGRES_DSN",
		"postgres://admin:password@localhost:5433/covid19?sslmode=disable",
	)
	db, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("could not ping database: %v", err)
	}
	return db, migrateDb(dsn)
}

func migrateDb(dsn string) error {
	pg, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open error: %s", err)
	}
	driver, err := postgres.WithInstance(pg, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("cannot init go-migrate: %s", err)
	}
	migrationsDir := env.EnvOrDefault("MIGRATIONS_DIR", "./migrations")
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("cannot create migrate.NewWithDatabaseInstance: %s", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("cannot migrate up: %s", err)
	}

	log.Println("all database migrations are complete")

	return nil
}
