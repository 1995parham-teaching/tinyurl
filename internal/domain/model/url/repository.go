package url

import (
	"context"
)

type Repository interface {
	Create(context.Context, URL) (string, error)
	Update(context.Context, URL) error
	FromShortURL(context.Context, string) (URL, error)
}
