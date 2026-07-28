package repository_test

import (
	"testing"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/infra/repository"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func newVisits(
	t *testing.T, cfg repository.VisitsConfig, next *fakeRepo,
) (*repository.BufferedVisits, *fxtest.Lifecycle) {
	t.Helper()

	lc := fxtest.NewLifecycle(t)
	visits := repository.NewBufferedVisits(cfg, lc, next, telemetry.ProvideNull(lc), zap.NewNop())

	return visits, lc
}

func slowFlushConfig() repository.VisitsConfig {
	return repository.VisitsConfig{
		Enabled: true,
		// long enough that nothing flushes on its own during a test
		FlushInterval: time.Hour,
		MaxBuffered:   1_000_000,
	}
}

// TestVisitsAreFolded is the reason the layer exists: a thousand redirects of the same link
// must cost one write, not a thousand writes against one row.
func TestVisitsAreFolded(t *testing.T) {
	t.Parallel()

	next := newFakeRepo()
	visits, _ := newVisits(t, slowFlushConfig(), next)

	for range 1000 {
		require.NoError(t, visits.IncrementVisits(t.Context(), "abc"))
	}

	require.Equal(t, 0, next.batchCount(), "counting a visit must not write")
	require.Equal(t, 1, visits.Pending())

	require.NoError(t, visits.Flush(t.Context()))

	require.Equal(t, 1, next.batchCount())
	require.Equal(t, uint64(1000), next.delta("abc"))
}

func TestVisitsFlushIsEmptyWhenNothingBuffered(t *testing.T) {
	t.Parallel()

	next := newFakeRepo()
	visits, _ := newVisits(t, slowFlushConfig(), next)

	require.NoError(t, visits.Flush(t.Context()))
	require.Equal(t, 0, next.batchCount(), "an empty buffer should not produce a statement")
}

// TestVisitsFlushOnInterval covers the background loop, which is what keeps counts moving when
// no single key is busy enough to trigger an early flush.
func TestVisitsFlushOnInterval(t *testing.T) {
	t.Parallel()

	cfg := slowFlushConfig()
	cfg.FlushInterval = 20 * time.Millisecond

	next := newFakeRepo()
	visits, lc := newVisits(t, cfg, next)

	lc.RequireStart()
	defer lc.RequireStop()

	require.NoError(t, visits.IncrementVisits(t.Context(), "abc"))

	require.Eventually(t, func() bool {
		return next.delta("abc") == 1
	}, time.Second, 5*time.Millisecond)
}

// TestVisitsFlushWhenFull keeps a burst from growing the buffer without bound between ticks.
func TestVisitsFlushWhenFull(t *testing.T) {
	t.Parallel()

	cfg := slowFlushConfig()
	cfg.MaxBuffered = 5

	next := newFakeRepo()
	visits, lc := newVisits(t, cfg, next)

	lc.RequireStart()
	defer lc.RequireStop()

	for i := range 5 {
		require.NoError(t, visits.IncrementVisits(t.Context(), string(rune('a'+i))))
	}

	require.Eventually(t, func() bool {
		return next.batchCount() >= 1
	}, time.Second, 5*time.Millisecond, "reaching the buffer limit should have triggered a flush")

	require.Equal(t, uint64(1), next.delta("a"))
	require.Equal(t, uint64(1), next.delta("e"))
}

// TestVisitsFlushOnShutdown covers the drain on the way out, without which a deploy would throw
// away every count gathered since the last tick.
func TestVisitsFlushOnShutdown(t *testing.T) {
	t.Parallel()

	next := newFakeRepo()
	visits, lc := newVisits(t, slowFlushConfig(), next)

	lc.RequireStart()

	require.NoError(t, visits.IncrementVisits(t.Context(), "abc"))
	require.Equal(t, uint64(0), next.delta("abc"))

	lc.RequireStop()

	require.Equal(t, uint64(1), next.delta("abc"), "buffered visits were lost on shutdown")
}

// TestVisitsSurviveAFailedFlush checks that a database blip delays counts rather than dropping
// them, which is the difference between a buffer and a leak.
func TestVisitsSurviveAFailedFlush(t *testing.T) {
	t.Parallel()

	next := newFakeRepo()
	visits, _ := newVisits(t, slowFlushConfig(), next)

	require.NoError(t, visits.IncrementVisits(t.Context(), "abc"))
	require.NoError(t, visits.IncrementVisits(t.Context(), "abc"))

	next.failNextBatch()

	require.Error(t, visits.Flush(t.Context()))
	require.Equal(t, 1, visits.Pending(), "counts should have been put back after the failure")

	// counts gathered while the write was failing must be added to those put back, not replace
	// them.
	require.NoError(t, visits.IncrementVisits(t.Context(), "abc"))

	require.NoError(t, visits.Flush(t.Context()))
	require.Equal(t, uint64(3), next.delta("abc"))
}
