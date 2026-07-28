package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

		// the primary key on urls.key is what actually guarantees short url uniqueness,
		// so a unique violation here means the key is taken and the caller must pick another.
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return urlrepo.ErrDuplicateShortURL
		}

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

	// expiry is filtered in the query rather than by the caller so that an expired url is
	// indistinguishable from a missing one and no call site can forget to check it.
	result, err := r.db.Where("key = ? AND (expire IS NULL OR expire > now())", key).First(ctx)
	if err != nil {
		r.logger.Error("fetching url from database failed", zap.Error(err), zap.String(logtag.Operation, "from-short-url"))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, urlrepo.ErrURLNotFound
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

// batchVisitsQuery folds a whole batch of counters into a single statement. The deltas travel
// as a VALUES list that is joined against urls, so the cost of recording a thousand visits is
// one round trip rather than a thousand. The casts are needed because postgres has no other
// way to know what type an otherwise untyped parameter of a VALUES list holds.
const batchVisitsQuery = `UPDATE urls
SET visits = COALESCE(urls.visits, 0) + visited.delta
FROM (VALUES %s) AS visited(key, delta)
WHERE urls.key = visited.key`

func (r *URLDB) IncrementVisitsBatch(ctx context.Context, deltas map[string]uint64) error {
	if len(deltas) == 0 {
		return nil
	}

	start := time.Now()

	// each row of the VALUES list contributes a key and a delta.
	const argsPerRow = 2

	rows := make([]string, 0, len(deltas))
	args := make([]any, 0, len(deltas)*argsPerRow)

	for key, delta := range deltas {
		rows = append(rows, "(?::text, ?::bigint)")
		args = append(args, key, delta)
	}

	query := fmt.Sprintf(batchVisitsQuery, strings.Join(rows, ", "))

	if err := r.db.Exec(ctx, query, args...); err != nil {
		r.logger.Error("incrementing visits in batch failed",
			zap.Error(err), zap.String(logtag.Operation, "increment-visits-batch"))

		return fmt.Errorf("incrementing visits in batch failed %w", err)
	}

	// rows that no longer exist simply do not match, which is not an error: a url may well be
	// removed between a visit being counted and the batch reaching the database.

	r.responseTime.Record(
		ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String(logtag.Operation, "increment-visits-batch"),
		),
	)

	return nil
}
