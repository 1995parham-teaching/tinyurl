package urlrepo

import (
	"context"
	"errors"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
)

var (
	ErrURLNotFound       = errors.New("url does not exist")
	ErrDuplicateShortURL = errors.New("short url already exists")
)

type Repository interface {
	Create(ctx context.Context, url url.URL) error
	Update(ctx context.Context, url url.URL) error
	FromShortURL(ctx context.Context, key string) (url.URL, error)
}
