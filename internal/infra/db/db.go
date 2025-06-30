package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	DB  *gorm.DB
	SQL *sql.DB
}

const PingTimeout = 10 * time.Second

func Provide(cfg Config, logger *zap.Logger) (*DB, error) {
	// nolint: exhaustruct
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		// enable prepared statements and caching them globally
		PrepareStmt: true,
		// gorm performs write operations insides a transaction to ensure data consistency.
		// which is bad for performance.
		SkipDefaultTransaction: false,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot open connection to the database %w", err)
	}

	logger.Info("open connection to the database successfully", zap.String("dsn", cfg.DSN))

	if cfg.Debug {
		db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("cannot get sql database from gorm %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), PingTimeout)
	defer cancel()

	err = sqlDB.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot ping the database %w", err)
	}

	logger.Info("get sql database from gorm successfully", zap.String("dsn", cfg.DSN))

	sqlDB.SetMaxIdleConns(cfg.MaxIdelConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &DB{DB: db, SQL: sqlDB}, nil
}
