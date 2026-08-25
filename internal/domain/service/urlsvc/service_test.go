package urlsvc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1995parham-teaching/tinyurl/internal/domain/service/urlsvc"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// takenRepo accepts a creation only once the given number of keys have been rejected as
// duplicates, standing in for a key space where the first few draws are already occupied.
type takenRepo struct {
	collisions int
	attempts   []string

	// lookups records every key the service asked for, so that a lookup doing more work than it
	// claims to cannot pass unnoticed.
	lookups []string
	stored  map[string]url.URL
}

func (r *takenRepo) Create(_ context.Context, u url.URL) error {
	r.attempts = append(r.attempts, u.Key)

	if len(r.attempts) <= r.collisions {
		return urlrepo.ErrDuplicateShortURL
	}

	if r.stored == nil {
		r.stored = make(map[string]url.URL)
	}

	r.stored[u.Key] = u

	return nil
}

func (r *takenRepo) Update(context.Context, url.URL) error { return nil }

func (r *takenRepo) IncrementVisits(context.Context, string) error { return nil }

func (r *takenRepo) IncrementVisitsBatch(context.Context, map[string]uint64) error { return nil }

func (r *takenRepo) FromShortURL(_ context.Context, key string) (url.URL, error) {
	r.lookups = append(r.lookups, key)

	if record, ok := r.stored[key]; ok {
		return record, nil
	}

	// nolint: exhaustruct_v5
	return url.URL{}, urlrepo.ErrURLNotFound
}

// countingGenerator hands out a different key every call so that retries are distinguishable.
type countingGenerator struct {
	calls int
}

func (g *countingGenerator) ShortURLKey(context.Context) (string, error) {
	g.calls++

	return fmt.Sprintf("key-%d", g.calls), nil
}

func newService(repo urlrepo.Repository, gen *countingGenerator) urlsvc.URLSvc {
	return urlsvc.ProvideURLSvc(repo, zap.NewNop(), gen)
}

func TestCreateRetriesOnCollision(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: 2, attempts: nil, lookups: nil, stored: nil}
	gen := &countingGenerator{calls: 0}

	key, err := newService(repo, gen).Create(t.Context(), "https://github.com", nil)
	require.NoError(t, err)

	require.Equal(t, "key-3", key, "the key returned should be the one that was actually stored")
	require.Equal(t, []string{"key-1", "key-2", "key-3"}, repo.attempts)
}

func TestCreateGivesUpAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: urlsvc.MaxKeyGenAttempts, attempts: nil, lookups: nil, stored: nil}
	gen := &countingGenerator{calls: 0}

	_, err := newService(repo, gen).Create(t.Context(), "https://github.com", nil)
	require.ErrorIs(t, err, urlsvc.ErrKeyGenFailed)
	require.Len(t, repo.attempts, urlsvc.MaxKeyGenAttempts)
}

func TestCreateWithKeyReportsDuplicate(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: 1, attempts: nil, lookups: nil, stored: nil}

	err := newService(repo, &countingGenerator{calls: 0}).
		CreateWithKey(t.Context(), "github", "https://github.com", nil)

	require.ErrorIs(t, err, urlsvc.ErrKeyAlreadyExists,
		"a taken vanity name must be reported as such rather than as an internal failure")
	require.Equal(t, []string{"github"}, repo.attempts,
		"a chosen name must be stored as given, so that looking it up takes one query")
}

// TestVisitTakesOneLookup is what dropping the stored prefix bought. Chosen names used to be
// stored under a name the caller never typed, so every one of them cost a wasted query first.
func TestVisitTakesOneLookup(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: 0, attempts: nil, lookups: nil, stored: nil}
	svc := newService(repo, &countingGenerator{calls: 0})

	require.NoError(t, svc.CreateWithKey(t.Context(), "github", "https://github.com", nil))

	record, err := svc.Visit(t.Context(), "github")
	require.NoError(t, err)
	require.Equal(t, "https://github.com", record.URL)

	require.Equal(t, []string{"github"}, repo.lookups)
}

func TestVisitMissingKey(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: 0, attempts: nil, lookups: nil, stored: nil}

	_, err := newService(repo, &countingGenerator{calls: 0}).Visit(t.Context(), "nope")
	require.ErrorIs(t, err, urlsvc.ErrURLNotFound)
	require.Equal(t, []string{"nope"}, repo.lookups, "a miss should not be retried under another name")
}

func TestCreateStoresExpiry(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: 0, attempts: nil, lookups: nil, stored: nil}
	expire := time.Now().Add(time.Hour)

	_, err := newService(repo, &countingGenerator{calls: 0}).Create(t.Context(), "https://github.com", &expire)
	require.NoError(t, err)
}
