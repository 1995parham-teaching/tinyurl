package server

import (
	"github.com/1989michael/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1989michael/tinyurl/internal/domain/service/urlsvc"
	"github.com/1989michael/tinyurl/internal/infra/config"
	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/1989michael/tinyurl/internal/infra/generator"
	"github.com/1989michael/tinyurl/internal/infra/http/server"
	"github.com/1989michael/tinyurl/internal/infra/logger"
	"github.com/1989michael/tinyurl/internal/infra/repository"
	"github.com/1989michael/tinyurl/internal/infra/telemetry"
	"github.com/labstack/echo/v4"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
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
			PersistentPreRun: func(_ *cobra.Command, _ []string) {
				pterm.DefaultCenter.Println("Shorten your URL to easily remember them and share them with your clients")

				s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("TinyURL")).Srender()
				pterm.DefaultCenter.Println(s)

				pterm.DefaultCenter.WithCenterEachLineSeparately().Println("Michael Weiss\nJuly 2023")
			},
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.Provide(telemetry.Provide),
					fx.Provide(db.Provide),
					fx.Provide(generator.Provide),
					fx.Provide(
						fx.Annotate(repository.ProvideURLDB, fx.As(new(urlrepo.Repository))),
					),
					fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
						return &fxevent.ZapLogger{Logger: logger}
					}),
					fx.Provide(urlsvc.ProvideURLSvc),
					fx.Provide(server.Provide),
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
