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
	Create(context.Context, url.URL) error
	Update(context.Context, url.URL) error
	FromShortURL(context.Context, string) (url.URL, error)
}
