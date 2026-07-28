package urlrepo

import (
	"context"
	"errors"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
)

var (
	ErrURLNotFound       = errors.New("url does not exist")
	ErrDuplicateShortURL = errors.New("short url already exists")
)

type Repository interface {
	Create(ctx context.Context, url url.URL) error
	Update(ctx context.Context, url url.URL) error
	FromShortURL(ctx context.Context, key string) (url.URL, error)
	IncrementVisits(ctx context.Context, key string) error

	// IncrementVisitsBatch records several visits at once. Redirects outnumber creations by
	// orders of magnitude, so counting them one statement at a time makes every read a write
	// against a single row; this is how those writes get folded together.
	IncrementVisitsBatch(ctx context.Context, deltas map[string]uint64) error
}
