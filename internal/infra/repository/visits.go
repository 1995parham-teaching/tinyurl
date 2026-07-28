package repository

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// BufferedVisits accumulates visit counts in memory and writes them out in batches.
//
// Recording a visit used to be an UPDATE against a single row on the redirect path, so the
// busiest links contended on their own row and every read cost a write. Here a visit only
// increments a counter under a mutex; a background goroutine folds the counters into one
// statement every FlushInterval, or sooner when too many keys have piled up.
//
// The trade is deliberate: a crash loses at most one interval of counts, and a counter read
// straight from the database lags by up to one interval. Neither matters for a visit tally, and
// both buy the redirect path its write back.
type BufferedVisits struct {
	next urlrepo.Repository
	cfg  VisitsConfig

	logger *zap.Logger

	mu      sync.Mutex
	pending map[string]uint64

	// trigger asks for an early flush. The loop runs under its own context rather than under a
	// request or startup one, because it outlives both; cancelling it is what stops the loop,
	// and done closes once it has returned, so the final flush cannot race with it.
	trigger chan struct{}
	stop    context.CancelFunc
	done    chan struct{}

	flushed metric.Int64Counter
}

func NewBufferedVisits(
	cfg VisitsConfig, lc fx.Lifecycle, next urlrepo.Repository, tele telemetry.Telemetery, logger *zap.Logger,
) *BufferedVisits {
	meter := tele.MeterProvider.Meter("repository.visits")

	flushed, err := meter.Int64Counter("visits.flushed")
	if err != nil {
		panic(err)
	}

	// the loop must outlive the start context, which is cancelled as soon as startup finishes,
	// so it gets one of its own. cancel is held on the struct and called from the OnStop hook.
	loopCtx, stop := context.WithCancel(context.Background()) // nolint: gosec

	r := &BufferedVisits{
		next:    next,
		cfg:     cfg,
		logger:  logger.Named("repository.visits"),
		mu:      sync.Mutex{},
		pending: make(map[string]uint64),
		trigger: make(chan struct{}, 1),
		stop:    stop,
		done:    make(chan struct{}),
		flushed: flushed,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go r.run(loopCtx)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			r.stop()
			<-r.done

			// whatever is still buffered belongs in the database, not in a dead process.
			if err := r.Flush(ctx); err != nil {
				r.logger.Error("flushing visits on shutdown failed", zap.Error(err))
			}

			return nil
		},
	})

	return r
}

// IncrementVisits records a visit in memory. It does not touch the database and so does not
// fail, which is why the redirect path can afford to call it on every request.
func (r *BufferedVisits) IncrementVisits(_ context.Context, key string) error {
	r.mu.Lock()
	r.pending[key]++
	full := len(r.pending) >= r.cfg.MaxBuffered
	r.mu.Unlock()

	if full {
		select {
		case r.trigger <- struct{}{}:
		default:
			// a flush is already pending, which is all this needs to achieve.
		}
	}

	return nil
}

// IncrementVisitsBatch passes through: it is the drain the buffer itself writes into.
func (r *BufferedVisits) IncrementVisitsBatch(ctx context.Context, deltas map[string]uint64) error {
	return r.next.IncrementVisitsBatch(ctx, deltas)
}

// Flush writes everything buffered so far. Counts are taken out of the buffer before the write
// and put back if it fails, so a database blip delays them rather than dropping them.
func (r *BufferedVisits) Flush(ctx context.Context) error {
	r.mu.Lock()

	if len(r.pending) == 0 {
		r.mu.Unlock()

		return nil
	}

	pending := r.pending
	r.pending = make(map[string]uint64, len(pending))

	r.mu.Unlock()

	if err := r.next.IncrementVisitsBatch(ctx, pending); err != nil {
		r.restore(pending)

		return fmt.Errorf("flushing buffered visits failed %w", err)
	}

	var total uint64
	for _, delta := range pending {
		total += delta
	}

	if total > math.MaxInt64 {
		total = math.MaxInt64
	}

	r.flushed.Add(ctx, int64(total))

	return nil
}

// Pending reports how many keys are waiting to be written, for tests and for diagnostics.
func (r *BufferedVisits) Pending() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.pending)
}

func (r *BufferedVisits) Create(ctx context.Context, u url.URL) error {
	return r.next.Create(ctx, u)
}

func (r *BufferedVisits) Update(ctx context.Context, u url.URL) error {
	return r.next.Update(ctx, u)
}

func (r *BufferedVisits) FromShortURL(ctx context.Context, key string) (url.URL, error) {
	return r.next.FromShortURL(ctx, key)
}

func (r *BufferedVisits) run(ctx context.Context) {
	defer close(r.done)

	ticker := time.NewTicker(r.cfg.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-r.trigger:
		}

		// the write is shielded from cancellation: it carries counts that several requests
		// contributed to, and shutdown should let it finish rather than abandon them.
		if err := r.Flush(context.WithoutCancel(ctx)); err != nil {
			r.logger.Error("flushing visits failed", zap.Error(err))
		}
	}
}

// restore folds counts back into the buffer after a failed write, adding to whatever arrived in
// the meantime rather than overwriting it.
func (r *BufferedVisits) restore(pending map[string]uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, delta := range pending {
		r.pending[key] += delta
	}
}
