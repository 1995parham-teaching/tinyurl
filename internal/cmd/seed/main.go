package seed

import (
	"context"

	"github.com/1989michael/tinyurl/internal/infra/config"
	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main(cfg config.Config, logger *zap.Logger) {
	db, err := db.New(cfg.Database)
	if err != nil {
		logger.Fatal("database initiation failed", zap.Error(err))
	}
}

// Register migrate command.
func Register(root *cobra.Command, cfg config.Config, logger *zap.Logger) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "seed",
			Short: "Add records into database",
			Run: func(_ *cobra.Command, _ []string) {
				main(cfg, logger)
			},
		},
	)
}
