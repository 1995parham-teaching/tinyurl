package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
	"github.com/1989michael/tinyurl/internal/domain/repourl"
)

var (
	ErrKeyGenFailed     = errors.New("cannot generate new random string as short url")
	ErrKeyAlreadyExists = errors.New("given static key already exists")
)

// shortURLKey generates a random key from the source characters.
func shortURLKey() string {
	const (
		length = 6
		source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	)

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(source))))
		if err != nil {
			panic(err)
		}

		b[i] = source[n.Int64()]
	}

	return string(b)
}

type URLSvc struct {
	repo repourl.Repository
}

func (s *URLSvc) create(ctx context.Context, key string, address string, expire *time.Time) error {
	url := url.URL{
		Key:    key,
		URL:    address,
		Visits: 0,
		Expire: expire,
	}

	if err := s.repo.Create(ctx, url); err != nil {
		return fmt.Errorf("url creation failed %w", err)
	}

	return nil
}

func (s *URLSvc) Create(ctx context.Context, address string, expire *time.Time) error {
	key := shortURLKey()

	if err := s.create(ctx, key, address, expire); err != nil {
		if errors.Is(err, repourl.ErrDuplicateShortURL) {
			return ErrKeyGenFailed
		}

		return fmt.Errorf("url creation failed %w", err)
	}

	return nil
}

func (s *URLSvc) CreateWithKey(ctx context.Context, key string, address string, expire *time.Time) error {
	key = fmt.Sprintf("static_%s", key)

	if err := s.create(ctx, key, address, expire); err != nil {
		if errors.Is(err, repourl.ErrDuplicateShortURL) {
			return ErrKeyAlreadyExists
		}

		return fmt.Errorf("url creation failed %w", err)
	}

	return nil
}
