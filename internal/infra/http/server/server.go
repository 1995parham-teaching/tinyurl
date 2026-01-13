package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/1995parham-teaching/tinyurl/internal/domain/service/urlsvc"
	"github.com/1995parham-teaching/tinyurl/internal/infra/http/handler"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Provide(
	lc fx.Lifecycle, cfg Config, logger *zap.Logger, tele telemetry.Telemetery, urlSvc urlsvc.URLSvc,
) *echo.Echo {
	app := echo.New()

	handler.Healthz{
		Logger: logger.Named("handler").Named("healthz"),
		Tracer: tele.TraceProvider.Tracer("handler.healthz"),
	}.Register(app.Group(""))

	handler.URL{
		Logger:  logger.Named("handler").Named("url"),
		Tracer:  tele.TraceProvider.Tracer("handler.url"),
		Service: urlSvc,
	}.Register(app.Group(""))

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				if err := app.Start(cfg.Address); !errors.Is(err, http.ErrServerClosed) {
					logger.Fatal("echo initiation failed", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: app.Shutdown,
	})

	return app
}
