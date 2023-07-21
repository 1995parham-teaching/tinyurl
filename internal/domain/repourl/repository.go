package repourl

import (
	"context"
	"errors"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
)

var ErrURLNotFound = errors.New("url does not exist")

type Repository interface {
	Create(context.Context, url.URL) error
	Update(context.Context, url.URL) error
	FromShortURL(context.Context, string) (url.URL, error)
}
