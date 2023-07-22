package migrate

import (
	"github.com/1989michael/tinyurl/internal/domain/model/url"
	"github.com/1989michael/tinyurl/internal/infra/config"
	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/1989michael/tinyurl/internal/infra/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main(logger *zap.Logger, db *db.DB, shutdonwer fx.Shutdowner) {
	logger.Info("running migrations using gorm")

	if err := db.DB.AutoMigrate(new(url.URL)); err != nil {
		logger.Fatal("migration failed", zap.Error(err))
	}

	logger.Info("migrations applied successfully")

	_ = shutdonwer.Shutdown()
}

// Register migrate command.
func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "migrate",
			Short: "Database migration",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.Provide(db.Provide),
					fx.NopLogger,
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
