package repository

import (
	"context"
	"fmt"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
	"github.com/1989michael/tinyurl/internal/infra/db"
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
	if err := r.db.DB.WithContext(ctx).Create(&url).Error; err != nil {
		return fmt.Errorf("url creation failed %w", err)
	}

	return nil
}

func (r *URLDB) GetWithShortURL() {
}

func (r *URLDB) Update() {
}
