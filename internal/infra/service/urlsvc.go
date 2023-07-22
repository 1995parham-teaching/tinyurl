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
	"go.uber.org/zap"
)

var (
	ErrKeyGenFailed     = errors.New("cannot generate new random string as short url")
	ErrKeyAlreadyExists = errors.New("given static key already exists")
	ErrURLNotFound      = repourl.ErrURLNotFound
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
	repo   repourl.Repository
	logger *zap.Logger
}

func ProvideURLSvc(repo repourl.Repository, logger *zap.Logger) *URLSvc {
	return &URLSvc{
		repo:   repo,
		logger: logger.Named("urlsvc"),
	}
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

func (s *URLSvc) Create(ctx context.Context, address string, expire *time.Time) (string, error) {
	key := shortURLKey()

	if err := s.create(ctx, key, address, expire); err != nil {
		if errors.Is(err, repourl.ErrDuplicateShortURL) {
			return "", ErrKeyGenFailed
		}

		return "", err
	}

	return key, nil
}

func (s *URLSvc) CreateWithKey(ctx context.Context, key string, address string, expire *time.Time) error {
	key = fmt.Sprintf("static_%s", key)

	if err := s.create(ctx, key, address, expire); err != nil {
		if errors.Is(err, repourl.ErrDuplicateShortURL) {
			return ErrKeyAlreadyExists
		}

		return err
	}

	return nil
}

func (s *URLSvc) visit(ctx context.Context, key string) (url.URL, error) {
	{
		url, err := s.repo.FromShortURL(ctx, key)
		if err != nil {
			if !errors.Is(err, repourl.ErrURLNotFound) {
				return url, fmt.Errorf("url fetching failed %w", err)
			}
		} else {
			return url, nil
		}
	}

	url, err := s.repo.FromShortURL(ctx, fmt.Sprintf("static_%s", key))
	if err != nil {
		if errors.Is(err, repourl.ErrURLNotFound) {
			return url, ErrURLNotFound
		}
		return url, fmt.Errorf("url fetching failed %w", err)
	}

	return url, nil
}

func (s *URLSvc) Visit(ctx context.Context, key string) (url.URL, error) {
	url, err := s.visit(ctx, key)
	if err != nil {
		return url, err
	}

	url.Visits += 1
	if err := s.repo.Update(ctx, url); err != nil {
		s.logger.Error("updating url visit coount failed", zap.Error(err))
	}

	return url, nil
}
