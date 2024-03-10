package app

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

const MigrationsRelDir = "deployment/postgres"

func createDBConnPool(ctx context.Context, dsn string) (pool *pgxpool.Pool, err error) {
	pgxConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return
	}
	pgxConfig.MaxConns = maxConnections
	pool, err = pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("pgx connection error: %w", err)
	}
	return pool, nil
}

func migrationsUP(ctx context.Context, pool *pgxpool.Pool, projRootPath string) (err error) {
	if err = goose.SetDialect("postgres"); err != nil {
		return
	}
	db := stdlib.OpenDBFromPool(pool)
	if err = goose.Up(db, projRootPath+"/"+MigrationsRelDir); err != nil {
		return
	}
	return nil
}
