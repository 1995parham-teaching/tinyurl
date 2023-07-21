package url

import (
	"context"
	"errors"
)

var ErrURLNotFound = errors.New("url does not exist")

type Repository interface {
	Create(context.Context, URL) error
	Update(context.Context, URL) error
	FromShortURL(context.Context, string) (URL, error)
}
