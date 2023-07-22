package server

import (
	"github.com/1989michael/tinyurl/internal/domain/repourl"
	"github.com/1989michael/tinyurl/internal/infra/config"
	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/1989michael/tinyurl/internal/infra/http/server"
	"github.com/1989michael/tinyurl/internal/infra/logger"
	"github.com/1989michael/tinyurl/internal/infra/repository"
	"github.com/1989michael/tinyurl/internal/infra/service"
	"github.com/1989michael/tinyurl/internal/infra/telemetry"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main(logger *zap.Logger, _ *echo.Echo) {
	logger.Info("welcome to our server")
}

// Register server command.
func Register(
	root *cobra.Command,
) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "server",
			Short: "Run server to serve the requests",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.Provide(telemetry.Provide),
					fx.Provide(db.Provide),
					fx.Provide(
						fx.Annotate(repository.ProvideURLDB, fx.As(new(repourl.Repository))),
					),
					fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
						return &fxevent.ZapLogger{Logger: logger}
					}),
					fx.Provide(service.ProvideURLSvc),
					fx.Provide(server.Provide),
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
