package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
	"github.com/1989michael/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/1989michael/tinyurl/internal/infra/logtag"
	"github.com/1989michael/tinyurl/internal/infra/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"gorm.io/gorm"
)

type URLDB struct {
	db *db.DB

	responeTime metric.Float64Histogram
}

func ProvideURLDB(db *db.DB, tele telemetry.Telemetery) *URLDB {
	meter := tele.MeterProvider.Meter("repository.urldb")

	rt, err := meter.Float64Histogram("response.time", metric.WithUnit("s"))
	if err != nil {
		panic(err)
	}

	return &URLDB{
		db:          db,
		responeTime: rt,
	}
}

func (r *URLDB) Create(ctx context.Context, url url.URL) error {
	start := time.Now()

	if err := r.db.DB.WithContext(ctx).Save(&url).Error; err != nil {
		return fmt.Errorf("url creation failed %w", err)
	}

	r.responeTime.Record(
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

	var url url.URL

	if err := r.db.DB.WithContext(ctx).Where("key = ?", key).First(&url).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return url, urlrepo.ErrURLNotFound
		}

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return url, urlrepo.ErrDuplicateShortURL
		}

		return url, fmt.Errorf("fetching url from database failed %w", err)
	}

	r.responeTime.Record(
		ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String(logtag.Operation, "from-short-url"),
		),
	)

	return url, nil
}

func (r *URLDB) Update(ctx context.Context, url url.URL) error {
	start := time.Now()

	if err := r.db.DB.WithContext(ctx).Save(url).Error; err != nil {
		return fmt.Errorf("updating url failed %w", err)
	}

	r.responeTime.Record(
		ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String(logtag.Operation, "update"),
		),
	)

	return nil
}
