package db

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	DB  *gorm.DB
	SQL *sql.DB
}

func New(cfg Config) (*DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot open connection to the database %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("cannot get sql database from gorm %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdelConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &DB{DB: db, SQL: sqlDB}, err
}
