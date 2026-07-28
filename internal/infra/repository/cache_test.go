package repository_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1995parham-teaching/tinyurl/internal/infra/repository"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func newCache(t *testing.T, cfg repository.CacheConfig, next *fakeRepo) *repository.Cached {
	t.Helper()

	return repository.NewCached(cfg, next, telemetry.ProvideNull(fxtest.NewLifecycle(t)), zap.NewNop())
}

func defaultCacheConfig() repository.CacheConfig {
	return repository.CacheConfig{
		Enabled:     true,
		Size:        100,
		TTL:         time.Minute,
		NegativeTTL: time.Minute,
	}
}

// TestCacheServesRepeatedReads is the point of the whole layer: a shortener answers the same
// handful of keys over and over, and only the first of those should reach the database.
func TestCacheServesRepeatedReads(t *testing.T) {
	t.Parallel()

	next := newFakeRepo()
	next.put("abc", testURL, sql.NullTime{Time: time.Time{}, Valid: false})

	cache := newCache(t, defaultCacheConfig(), next)

	for range 10 {
		record, err := cache.FromShortURL(t.Context(), "abc")
		require.NoError(t, err)
		require.Equal(t, testURL, record.URL)
	}

	require.Equal(t, 1, next.readCount(), "only the first read should have reached the database")
}

// TestCacheRemembersMisses protects the database from a scan for keys that do not exist, which
// short guessable keys invite.
func TestCacheRemembersMisses(t *testing.T) {
	t.Parallel()

	next := newFakeRepo()
	cache := newCache(t, defaultCacheConfig(), next)

	for range 10 {
		_, err := cache.FromShortURL(t.Context(), "nope")
		require.ErrorIs(t, err, urlrepo.ErrURLNotFound)
	}

	require.Equal(t, 1, next.readCount(), "only the first miss should have reached the database")
}

// TestCacheForgetsMissesQuickly bounds how long a negative entry can hide a key that has since
// been created somewhere else.
func TestCacheForgetsMissesQuickly(t *testing.T) {
	t.Parallel()

	cfg := defaultCacheConfig()
	cfg.NegativeTTL = 10 * time.Millisecond

	next := newFakeRepo()
	cache := newCache(t, cfg, next)

	_, err := cache.FromShortURL(t.Context(), "soon")
	require.ErrorIs(t, err, urlrepo.ErrURLNotFound)

	next.put("soon", testURL, sql.NullTime{Time: time.Time{}, Valid: false})

	require.Eventually(t, func() bool {
		_, err := cache.FromShortURL(t.Context(), "soon")

		return err == nil
	}, time.Second, 5*time.Millisecond)
}

// TestCacheCreateClearsMiss covers the common flow of claiming a vanity name and following it
// immediately: the miss recorded a moment earlier must not outlive the creation.
func TestCacheCreateClearsMiss(t *testing.T) {
	t.Parallel()

	next := newFakeRepo()
	cache := newCache(t, defaultCacheConfig(), next)

	_, err := cache.FromShortURL(t.Context(), "gh")
	require.ErrorIs(t, err, urlrepo.ErrURLNotFound)

	// nolint: exhaustruct
	require.NoError(t, cache.Create(t.Context(), url.URL{
		Key:    "gh",
		URL:    testURL,
		Visits: 0,
		Expire: sql.NullTime{Time: time.Time{}, Valid: false},
	}))

	record, err := cache.FromShortURL(t.Context(), "gh")
	require.NoError(t, err)
	require.Equal(t, testURL, record.URL)
}

// TestCacheDoesNotOutliveExpiry is the correctness constraint on the whole layer: the database
// refuses to return an expired url, and the cache must not undo that by holding one longer.
func TestCacheDoesNotOutliveExpiry(t *testing.T) {
	t.Parallel()

	cfg := defaultCacheConfig()
	cfg.TTL = time.Hour

	next := newFakeRepo()
	next.put("brief", testURL, sql.NullTime{
		Time:  time.Now().Add(20 * time.Millisecond),
		Valid: true,
	})

	cache := newCache(t, cfg, next)

	_, err := cache.FromShortURL(t.Context(), "brief")
	require.NoError(t, err, "the url should be served while it is still valid")

	require.Eventually(t, func() bool {
		_, err := cache.FromShortURL(t.Context(), "brief")

		return errors.Is(err, urlrepo.ErrURLNotFound)
	}, time.Second, 5*time.Millisecond, "an expired url was still being served from the cache")
}

// TestCacheEvictsBySize keeps a long tail of one-off keys from growing the process without
// bound.
func TestCacheEvictsBySize(t *testing.T) {
	t.Parallel()

	cfg := defaultCacheConfig()
	cfg.Size = 8

	next := newFakeRepo()
	cache := newCache(t, cfg, next)

	for i := range 100 {
		_, _ = cache.FromShortURL(t.Context(), string(rune('a'+i%26))+string(rune('a'+i)))
	}

	require.LessOrEqual(t, cache.Len(), cfg.Size)
}

// TestCachePassesVisitsThrough pins that the cache only stands in front of reads. Swallowing a
// visit here would silently stop the counter, and nothing else in the stack would notice.
func TestCachePassesVisitsThrough(t *testing.T) {
	t.Parallel()

	next := newFakeRepo()
	cache := newCache(t, defaultCacheConfig(), next)

	require.NoError(t, cache.IncrementVisits(t.Context(), "abc"))
	require.NoError(t, cache.IncrementVisitsBatch(t.Context(), map[string]uint64{"abc": 4}))

	require.Equal(t, uint64(5), next.delta("abc"))
}
