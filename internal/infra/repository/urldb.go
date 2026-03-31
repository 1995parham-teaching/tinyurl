package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
	"github.com/1995parham-teaching/tinyurl/internal/infra/logtag"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type URLDB struct {
	db     gorm.Interface[url.URL]
	logger *zap.Logger

	responseTime metric.Float64Histogram
}

func ProvideURLDB(db *db.DB, tele telemetry.Telemetery, logger *zap.Logger) *URLDB {
	meter := tele.MeterProvider.Meter("repository.urldb")

	rt, err := meter.Float64Histogram("response.time", metric.WithUnit("s"))
	if err != nil {
		panic(err)
	}

	return &URLDB{
		db:           gorm.G[url.URL](db.DB),
		responseTime: rt,
		logger:       logger.Named("repository.urldb"),
	}
}

func (r *URLDB) Create(ctx context.Context, u url.URL) error {
	start := time.Now()

	if err := r.db.Create(ctx, &u); err != nil {
		r.logger.Error("url creation failed", zap.Error(err), zap.String(logtag.Operation, "create"))

		return fmt.Errorf("url creation failed %w", err)
	}

	r.responseTime.Record(
		ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String(logtag.Operation, "create"),
		),
	)

	return nil
}

func (r *URLDB) FromShortURL(ctx context.Context, key string) (url.URL, error) {
	start := time.Now()

	result, err := r.db.Where("key = ?", key).First(ctx)
	if err != nil {
		r.logger.Error("fetching url from database failed", zap.Error(err), zap.String(logtag.Operation, "from-short-url"))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, urlrepo.ErrURLNotFound
		}

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return result, urlrepo.ErrDuplicateShortURL
		}

		return result, fmt.Errorf("fetching url from database failed %w", err)
	}

	r.responseTime.Record(
		ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String(logtag.Operation, "from-short-url"),
		),
	)

	return result, nil
}

func (r *URLDB) Update(ctx context.Context, u url.URL) error {
	start := time.Now()

	if _, err := r.db.Where("key = ?", u.Key).Updates(ctx, u); err != nil {
		r.logger.Error("updating url failed", zap.Error(err), zap.String(logtag.Operation, "update"))

		return fmt.Errorf("updating url failed %w", err)
	}

	r.responseTime.Record(
		ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String(logtag.Operation, "update"),
		),
	)

	return nil
}

func (r *URLDB) IncrementVisits(ctx context.Context, key string) error {
	start := time.Now()

	rowsAffected, err := r.db.Where("key = ?", key).Update(ctx, "visits", gorm.Expr("visits + ?", 1))
	if err != nil {
		r.logger.Error("incrementing visits failed",
			zap.Error(err), zap.String(logtag.Operation, "increment-visits"))

		return fmt.Errorf("incrementing visits failed %w", err)
	}

	if rowsAffected == 0 {
		return urlrepo.ErrURLNotFound
	}

	r.responseTime.Record(
		ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String(logtag.Operation, "increment-visits"),
		),
	)

	return nil
}
