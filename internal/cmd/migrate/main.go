package migrate

import (
	"context"
	"fmt"

	"github.com/1995parham-teaching/tinyurl/internal/infra/config"
	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
	"github.com/1995parham-teaching/tinyurl/internal/infra/logger"
	"github.com/1995parham-teaching/tinyurl/internal/infra/migrations"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Register migrate commands.
func Register(root *cobra.Command) {
	migrateCmd := &cobra.Command{ // nolint: exhaustruct
		Use:   "migrate",
		Short: "Database migration commands using Goose",
	}

	migrateCmd.AddCommand(
		&cobra.Command{ // nolint: exhaustruct
			Use:   "up",
			Short: "Migrate the DB to the most recent version available",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateUp)
			},
		},
		&cobra.Command{ // nolint: exhaustruct
			Use:   "up-by-one",
			Short: "Migrate the DB up by 1",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateUpByOne)
			},
		},
		&cobra.Command{ // nolint: exhaustruct
			Use:   "down",
			Short: "Roll back the version by 1",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateDown)
			},
		},
		&cobra.Command{ // nolint: exhaustruct
			Use:   "reset",
			Short: "Roll back all migrations",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateReset)
			},
		},
		&cobra.Command{ // nolint: exhaustruct
			Use:   "status",
			Short: "Dump the migration status for the current DB",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateStatus)
			},
		},
		&cobra.Command{ // nolint: exhaustruct
			Use:   "version",
			Short: "Print the current version of the database",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateVersion)
			},
		},
	)

	root.AddCommand(migrateCmd)
}

type migrationFunc func(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner)

func runMigration(fn migrationFunc) {
	fx.New(
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(db.Provide),
		fx.NopLogger,
		fx.Invoke(fn),
	).Run()
}

func newGooseProvider(database *db.DB) (*goose.Provider, error) {
	provider, err := goose.NewProvider(goose.DialectPostgres, database.SQL, migrations.FS)
	if err != nil {
		return nil, fmt.Errorf("failed to create goose provider: %w", err)
	}

	return provider, nil
}

func closeGooseProvider(logger *zap.Logger, provider *goose.Provider) {
	if err := provider.Close(); err != nil {
		logger.Error("failed to close goose provider", zap.Error(err))
	}
}

func migrateUp(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	provider, err := newGooseProvider(database)
	if err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}
	defer closeGooseProvider(logger, provider)

	results, err := provider.Up(context.Background())
	if err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	for _, result := range results {
		logger.Info(result.String())
	}

	logger.Info("migrations completed successfully")

	if err := shutdowner.Shutdown(); err != nil {
		logger.Error("shutdown failed", zap.Error(err))
	}
}

func migrateUpByOne(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	provider, err := newGooseProvider(database)
	if err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}
	defer closeGooseProvider(logger, provider)

	result, err := provider.UpByOne(context.Background())
	if err != nil {
		logger.Fatal("failed to run migration", zap.Error(err))
	}

	logger.Info(result.String())
	logger.Info("migration completed successfully")

	if err := shutdowner.Shutdown(); err != nil {
		logger.Error("shutdown failed", zap.Error(err))
	}
}

func migrateDown(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	provider, err := newGooseProvider(database)
	if err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}
	defer closeGooseProvider(logger, provider)

	result, err := provider.Down(context.Background())
	if err != nil {
		logger.Fatal("failed to rollback migration", zap.Error(err))
	}

	logger.Info(result.String())
	logger.Info("migration rolled back successfully")

	if err := shutdowner.Shutdown(); err != nil {
		logger.Error("shutdown failed", zap.Error(err))
	}
}

func migrateReset(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	provider, err := newGooseProvider(database)
	if err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}
	defer closeGooseProvider(logger, provider)

	results, err := provider.DownTo(context.Background(), 0)
	if err != nil {
		logger.Fatal("failed to reset migrations", zap.Error(err))
	}

	for _, result := range results {
		logger.Info(result.String())
	}

	logger.Info("all migrations rolled back successfully")

	if err := shutdowner.Shutdown(); err != nil {
		logger.Error("shutdown failed", zap.Error(err))
	}
}

func migrateStatus(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	provider, err := newGooseProvider(database)
	if err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}
	defer closeGooseProvider(logger, provider)

	statuses, err := provider.Status(context.Background())
	if err != nil {
		logger.Fatal("failed to get migration status", zap.Error(err))
	}

	for _, status := range statuses {
		logger.Info("migration status",
			zap.String("source", status.Source.Path),
			zap.String("state", string(status.State)),
			zap.Time("applied_at", status.AppliedAt),
		)
	}

	if err := shutdowner.Shutdown(); err != nil {
		logger.Error("shutdown failed", zap.Error(err))
	}
}

func migrateVersion(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	provider, err := newGooseProvider(database)
	if err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}
	defer closeGooseProvider(logger, provider)

	version, err := provider.GetDBVersion(context.Background())
	if err != nil {
		logger.Fatal("failed to get database version", zap.Error(err))
	}

	logger.Info("current database version", zap.Int64("version", version))

	if err := shutdowner.Shutdown(); err != nil {
		logger.Error("shutdown failed", zap.Error(err))
	}
}
