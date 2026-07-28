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
}

func (r *takenRepo) Create(_ context.Context, u url.URL) error {
	r.attempts = append(r.attempts, u.Key)

	if len(r.attempts) <= r.collisions {
		return urlrepo.ErrDuplicateShortURL
	}

	return nil
}

func (r *takenRepo) Update(context.Context, url.URL) error { return nil }

func (r *takenRepo) IncrementVisits(context.Context, string) error { return nil }

func (r *takenRepo) FromShortURL(context.Context, string) (url.URL, error) {
	// nolint: exhaustruct
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

	repo := &takenRepo{collisions: 2, attempts: nil}
	gen := &countingGenerator{calls: 0}

	key, err := newService(repo, gen).Create(t.Context(), "https://github.com", nil)
	require.NoError(t, err)

	require.Equal(t, "key-3", key, "the key returned should be the one that was actually stored")
	require.Equal(t, []string{"key-1", "key-2", "key-3"}, repo.attempts)
}

func TestCreateGivesUpAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: urlsvc.MaxKeyGenAttempts, attempts: nil}
	gen := &countingGenerator{calls: 0}

	_, err := newService(repo, gen).Create(t.Context(), "https://github.com", nil)
	require.ErrorIs(t, err, urlsvc.ErrKeyGenFailed)
	require.Len(t, repo.attempts, urlsvc.MaxKeyGenAttempts)
}

func TestCreateWithKeyReportsDuplicate(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: 1, attempts: nil}

	err := newService(repo, &countingGenerator{calls: 0}).
		CreateWithKey(t.Context(), "github", "https://github.com", nil)

	require.ErrorIs(t, err, urlsvc.ErrKeyAlreadyExists,
		"a taken vanity name must be reported as such rather than as an internal failure")
	require.Equal(t, []string{"static_github"}, repo.attempts)
}

func TestCreateStoresExpiry(t *testing.T) {
	t.Parallel()

	repo := &takenRepo{collisions: 0, attempts: nil}
	expire := time.Now().Add(time.Hour)

	_, err := newService(repo, &countingGenerator{calls: 0}).Create(t.Context(), "https://github.com", &expire)
	require.NoError(t, err)
}
