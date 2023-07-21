package repository

import (
	"context"

	"github.com/1989michael/tinyurl/internal/domain/model"
)

type URL interface {
	Create(context.Context, model.URL) (string, error)
	Update(context.Context, model.URL) error
	FromShortURL(context.Context, string) (model.URL, error)
}
