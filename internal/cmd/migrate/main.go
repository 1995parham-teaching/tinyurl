package migrate

import (
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

const dialect = "postgres"

// Register migrate commands.
func Register(root *cobra.Command) {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands using Goose",
	}

	migrateCmd.AddCommand(
		&cobra.Command{
			Use:   "up",
			Short: "Migrate the DB to the most recent version available",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateUp)
			},
		},
		&cobra.Command{
			Use:   "up-by-one",
			Short: "Migrate the DB up by 1",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateUpByOne)
			},
		},
		&cobra.Command{
			Use:   "down",
			Short: "Roll back the version by 1",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateDown)
			},
		},
		&cobra.Command{
			Use:   "reset",
			Short: "Roll back all migrations",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateReset)
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Dump the migration status for the current DB",
			Run: func(_ *cobra.Command, _ []string) {
				runMigration(migrateStatus)
			},
		},
		&cobra.Command{
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

func setupGoose(logger *zap.Logger, database *db.DB) error {
	logger.Info("setting up goose migration")

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	return nil
}

func migrateUp(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	if err := setupGoose(logger, database); err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}

	if err := goose.Up(database.SQL, "."); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	logger.Info("migrations completed successfully")

	_ = shutdowner.Shutdown()
}

func migrateUpByOne(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	if err := setupGoose(logger, database); err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}

	if err := goose.UpByOne(database.SQL, "."); err != nil {
		logger.Fatal("failed to run migration", zap.Error(err))
	}

	logger.Info("migration completed successfully")

	_ = shutdowner.Shutdown()
}

func migrateDown(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	if err := setupGoose(logger, database); err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}

	if err := goose.Down(database.SQL, "."); err != nil {
		logger.Fatal("failed to rollback migration", zap.Error(err))
	}

	logger.Info("migration rolled back successfully")

	_ = shutdowner.Shutdown()
}

func migrateReset(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	if err := setupGoose(logger, database); err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}

	if err := goose.Reset(database.SQL, "."); err != nil {
		logger.Fatal("failed to reset migrations", zap.Error(err))
	}

	logger.Info("all migrations rolled back successfully")

	_ = shutdowner.Shutdown()
}

func migrateStatus(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	if err := setupGoose(logger, database); err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}

	if err := goose.Status(database.SQL, "."); err != nil {
		logger.Fatal("failed to get migration status", zap.Error(err))
	}

	_ = shutdowner.Shutdown()
}

func migrateVersion(logger *zap.Logger, database *db.DB, shutdowner fx.Shutdowner) {
	if err := setupGoose(logger, database); err != nil {
		logger.Fatal("failed to setup goose", zap.Error(err))
	}

	version, err := goose.GetDBVersion(database.SQL)
	if err != nil {
		logger.Fatal("failed to get database version", zap.Error(err))
	}

	logger.Info("current database version", zap.Int64("version", version))

	_ = shutdowner.Shutdown()
}
