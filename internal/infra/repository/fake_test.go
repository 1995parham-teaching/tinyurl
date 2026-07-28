package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
)

// errUnavailable stands in for a database that is not answering.
var errUnavailable = errors.New("database is unavailable")

// fakeRepo is an in-memory repository that counts what reached it, so the decorators can be
// tested by what they let through rather than by what they claim.
type fakeRepo struct {
	mu sync.Mutex

	records map[string]url.URL

	reads    int
	batches  int
	deltas   map[string]uint64
	failNext bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		mu:       sync.Mutex{},
		records:  make(map[string]url.URL),
		reads:    0,
		batches:  0,
		deltas:   make(map[string]uint64),
		failNext: false,
	}
}

func (r *fakeRepo) Create(_ context.Context, u url.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.records[u.Key]; ok {
		return urlrepo.ErrDuplicateShortURL
	}

	r.records[u.Key] = u

	return nil
}

func (r *fakeRepo) Update(_ context.Context, u url.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.records[u.Key] = u

	return nil
}

func (r *fakeRepo) FromShortURL(_ context.Context, key string) (url.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reads++

	record, ok := r.records[key]
	if !ok {
		// nolint: exhaustruct
		return url.URL{}, urlrepo.ErrURLNotFound
	}

	// the database refuses to return an expired url, and so must the fake, otherwise the cache
	// would look correct while serving something the database would have withheld.
	if record.Expire.Valid && time.Now().After(record.Expire.Time) {
		// nolint: exhaustruct
		return url.URL{}, urlrepo.ErrURLNotFound
	}

	return record, nil
}

func (r *fakeRepo) IncrementVisits(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.deltas[key]++

	return nil
}

func (r *fakeRepo) IncrementVisitsBatch(_ context.Context, deltas map[string]uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.failNext {
		r.failNext = false

		return errUnavailable
	}

	r.batches++

	for key, delta := range deltas {
		r.deltas[key] += delta
	}

	return nil
}

func (r *fakeRepo) put(key, address string, expire sql.NullTime) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// nolint: exhaustruct
	r.records[key] = url.URL{Key: key, URL: address, Visits: 0, Expire: expire}
}

func (r *fakeRepo) readCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.reads
}

func (r *fakeRepo) batchCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.batches
}

func (r *fakeRepo) delta(key string) uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.deltas[key]
}

func (r *fakeRepo) failNextBatch() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.failNext = true
}
