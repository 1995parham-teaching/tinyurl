package seed

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
	"github.com/1989michael/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1989michael/tinyurl/internal/infra/config"
	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/1989michael/tinyurl/internal/infra/logger"
	"github.com/1989michael/tinyurl/internal/infra/repository"
	"github.com/1989michael/tinyurl/internal/infra/telemetry"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main(logger *zap.Logger, repo urlrepo.Repository, shutdowner fx.Shutdowner) {
	ctx := context.Background()

	records := []url.URL{
		{
			Key:    "static_google",
			URL:    "https://google.com",
			Visits: 0,
			Expire: sql.NullTime{
				Time:  time.Now(),
				Valid: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, record := range records {
		if err := repo.Create(ctx, record); err != nil {
			if errors.Is(err, urlrepo.ErrDuplicateShortURL) {
				continue
			}

			logger.Fatal("cannot create record", zap.Error(err))
		}
	}

	_ = shutdowner.Shutdown()
}

// Register seed command.
func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "seed",
			Short: "Add sample records into database",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.Provide(db.Provide),
					fx.Provide(telemetry.ProvideNull),
					fx.Provide(
						fx.Annotate(repository.ProvideURLDB, fx.As(new(urlrepo.Repository))),
					),
					fx.NopLogger,
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
