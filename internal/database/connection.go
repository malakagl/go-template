package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/malakagl/kart-challenge/internal/config"
	"github.com/malakagl/kart-challenge/pkg/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB
var mu sync.Mutex

func Connect(ctx context.Context, cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Ensure singleton
	mu.Lock()
	defer mu.Unlock()
	if dbInstance != nil {
		log.WithCtx(ctx).Debug().Msg("Reusing existing GORM database connection")
		return dbInstance, nil
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.MaxConnMaxLifeTime)
	log.WithCtx(ctx).Debug().Msg("Connected to PostgreSQL via GORM")

	dbInstance = db

	return dbInstance, nil
}
