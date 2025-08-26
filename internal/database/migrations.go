package database

import (
	"context"
	"fmt"
	"regexp"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/malakagl/kart-challenge/internal/config"
	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/otel"
)

func RunMigrations(ctx context.Context, cfg *config.DatabaseConfig) error {
	spanCtx, span := otel.Tracer(ctx, "runMigrations")
	defer span.End()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)
	re := regexp.MustCompile(`:(.*?)@`)
	safeDsn := re.ReplaceAllString(dsn, ":*****@")
	log.WithCtx(ctx).Debug().Msgf("database migrations started: %v", safeDsn)
	migrationSource := fmt.Sprintf("file://%s/migrations", cfg.MigrationsFolderPath)
	m, err := migrate.New(migrationSource, dsn)
	if err != nil {
		log.WithCtx(ctx).Error().Msgf("Failed to create migration instance: %v - migrationsSource: %s", err, migrationSource)
		span.RecordError(err)
		return err
	}

	defer m.Close()
	if err2 := m.Up(); err2 != nil && !errors.Is(err2, migrate.ErrNoChange) {
		log.WithCtx(spanCtx).Error().Msgf("Migration failed: %v", err2)
		span.RecordError(err2)
		return err2
	}

	return nil
}
