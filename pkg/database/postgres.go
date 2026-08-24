package database

import (
	"context"
	"database/sql"
	"log"
	"time"
	"trash/api/pkg/config"
)

func InitPostgres(config config.Config) *sql.DB {
	db, err := sql.Open("postgres", config.GetDSN())
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	return db
}
