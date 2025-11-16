package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	_ "github.com/lib/pq"

	"github.com/exPriceD/pr-reviewer-service/internal/infrastructure/config"
)

type PostgresDB struct {
	db     *sql.DB
	getter *trmsql.CtxGetter
}

func NewPostgresDB(cfg config.DatabaseConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.PingTimeout)*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	getter := trmsql.DefaultCtxGetter

	return &PostgresDB{
		db:     db,
		getter: getter,
	}, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

// Getter возвращает CtxGetter для получения транзакции из контекста
func (p *PostgresDB) Getter() *trmsql.CtxGetter {
	return p.getter
}
