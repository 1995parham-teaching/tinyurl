package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
	"github.com/1989michael/tinyurl/internal/infra/db"
	"gorm.io/gorm"
)

type URLDB struct {
	db *db.DB
}

func NewURLDB(db *db.DB) *URLDB {
	return &URLDB{
		db: db,
	}
}

func (r *URLDB) Create(ctx context.Context, url url.URL) error {
	if err := r.db.DB.WithContext(ctx).Save(&url).Error; err != nil {
		return fmt.Errorf("url creation failed %w", err)
	}

	return nil
}

func (r *URLDB) FromShortURL(ctx context.Context, key string) (url.URL, error) {
	var instance url.URL

	if err := r.db.DB.WithContext(ctx).Where("key = ?", key).First(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return instance, url.ErrURLNotFound
		}

		return instance, fmt.Errorf("fetching url from database failed %w", err)
	}

	return instance, nil
}

func (r *URLDB) Update(ctx context.Context, url url.URL) error {
	if err := r.db.DB.WithContext(ctx).Save(url).Error; err != nil {
		return fmt.Errorf("updating url failed %w", err)
	}

	return nil
}
