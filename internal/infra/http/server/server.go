package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/service/urlsvc"
	"github.com/1995parham-teaching/tinyurl/internal/infra/http/handler"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// gracefulTimeout is how long StartConfig waits for in-flight requests to
// finish once a shutdown is requested.
const gracefulTimeout = 10 * time.Second

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

	// v5 removed Echo.Shutdown; graceful shutdown is driven by cancelling the
	// context passed to StartConfig.Start. We bridge that into fx's lifecycle:
	// cancel is invoked from the OnStop hook below.
	srvCtx, cancel := context.WithCancel(context.Background()) // nolint: gosec
	done := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				defer close(done)

				// nolint: exhaustruct
				sc := echo.StartConfig{
					Address:         cfg.Address,
					GracefulTimeout: gracefulTimeout,
				}

				if err := sc.Start(srvCtx, app); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Fatal("echo initiation failed", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			<-done

			return nil
		},
	})

	return app
}
