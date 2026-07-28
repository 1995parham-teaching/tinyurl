package repository

import (
	"context"
	"errors"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1995parham-teaching/tinyurl/internal/infra/logtag"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// entry is what the cache holds. A miss is cached too, so found records whether this entry
// stands for a url or for the absence of one.
type entry struct {
	url   url.URL
	found bool

	// deadline is when this entry stops being trusted. It is tracked per entry rather than
	// left to the cache's own expiry because an entry may need to lapse earlier than the
	// configured TTL, namely when the url it holds expires first.
	deadline time.Time
}

// Cached answers lookups from memory where it can.
//
// It is a decorator: it satisfies the same repository interface as the layer beneath it and can
// be left out of the stack entirely without anything else noticing. Only reads are served from
// memory; writes always go through, and invalidate the key they touch.
//
// The cache is per instance. Two consequences worth naming: a negative entry keeps one instance
// from seeing a key another instance just created until it lapses, and the visit counts it hands
// back are whatever they were when the entry was filled. Nothing serves those counts to a
// caller, but they should not be mistaken for current.
type Cached struct {
	next urlrepo.Repository
	cfg  CacheConfig

	entries *lru.LRU[string, entry]
	logger  *zap.Logger

	lookups metric.Int64Counter
}

func NewCached(
	cfg CacheConfig, next urlrepo.Repository, tele telemetry.Telemetery, logger *zap.Logger,
) *Cached {
	meter := tele.MeterProvider.Meter("repository.cache")

	lookups, err := meter.Int64Counter("cache.lookups")
	if err != nil {
		panic(err)
	}

	return &Cached{
		next:    next,
		cfg:     cfg,
		entries: lru.NewLRU[string, entry](cfg.Size, nil, cfg.TTL),
		logger:  logger.Named("repository.cache"),
		lookups: lookups,
	}
}

func (r *Cached) FromShortURL(ctx context.Context, key string) (url.URL, error) {
	if cached, ok := r.load(key); ok {
		r.record(ctx, "hit")

		if !cached.found {
			// nolint: exhaustruct
			return url.URL{}, urlrepo.ErrURLNotFound
		}

		return cached.url, nil
	}

	r.record(ctx, "miss")

	record, err := r.next.FromShortURL(ctx, key)
	if err != nil {
		if errors.Is(err, urlrepo.ErrURLNotFound) {
			r.store(key, entry{url: record, found: false, deadline: time.Now().Add(r.cfg.NegativeTTL)})
		}

		return record, err
	}

	r.store(key, entry{url: record, found: true, deadline: r.deadline(record)})

	return record, nil
}

func (r *Cached) Create(ctx context.Context, u url.URL) error {
	// dropping the entry before the write means a failed write cannot leave a stale one behind,
	// and clears any negative entry so a freshly created key resolves right away.
	r.entries.Remove(u.Key)

	if err := r.next.Create(ctx, u); err != nil {
		return err
	}

	r.entries.Remove(u.Key)

	return nil
}

func (r *Cached) Update(ctx context.Context, u url.URL) error {
	r.entries.Remove(u.Key)

	if err := r.next.Update(ctx, u); err != nil {
		return err
	}

	r.entries.Remove(u.Key)

	return nil
}

// IncrementVisits passes straight through. Visit counts are not what the cache is protecting,
// and the layer beneath is where they are folded together.
func (r *Cached) IncrementVisits(ctx context.Context, key string) error {
	return r.next.IncrementVisits(ctx, key)
}

func (r *Cached) IncrementVisitsBatch(ctx context.Context, deltas map[string]uint64) error {
	return r.next.IncrementVisitsBatch(ctx, deltas)
}

// Len reports how many entries are held, for tests and for diagnostics.
func (r *Cached) Len() int {
	return r.entries.Len()
}

func (r *Cached) load(key string) (entry, bool) {
	cached, ok := r.entries.Get(key)
	if !ok {
		return cached, false
	}

	if time.Now().After(cached.deadline) {
		r.entries.Remove(key)

		return cached, false
	}

	return cached, true
}

func (r *Cached) store(key string, e entry) {
	r.entries.Add(key, e)
}

// deadline is the configured TTL, brought forward when the url expires before it would lapse,
// so that an expired url is never served out of memory after the database stopped returning it.
func (r *Cached) deadline(record url.URL) time.Time {
	deadline := time.Now().Add(r.cfg.TTL)

	if record.Expire.Valid && record.Expire.Time.Before(deadline) {
		return record.Expire.Time
	}

	return deadline
}

func (r *Cached) record(ctx context.Context, result string) {
	r.lookups.Add(ctx, 1, metric.WithAttributes(
		attribute.String(logtag.Operation, "from-short-url"),
		attribute.String("result", result),
	))
}
