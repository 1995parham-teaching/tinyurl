package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/1989michael/tinyurl/internal/infra/http/handler"
	"github.com/1989michael/tinyurl/internal/infra/telemetry"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Provide(lc fx.Lifecycle, logger *zap.Logger, tele telemetry.Telemetery) *echo.Echo {
	app := echo.New()

	handler.Healthz{
		Logger: logger.Named("handler").Named("healthz"),
		Tracer: tele.TraceProvider.Tracer("handler.healthz"),
	}.Register(app.Group(""))

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				if err := app.Start(":1378"); !errors.Is(err, http.ErrServerClosed) {
					logger.Fatal("echo initiation failed", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown(ctx)
		},
	})

	return app
}
