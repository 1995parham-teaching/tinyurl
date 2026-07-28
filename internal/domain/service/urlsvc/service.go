package urlsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"go.uber.org/zap"
)

var (
	ErrKeyGenFailed     = errors.New("cannot generate new random string as short url")
	ErrKeyAlreadyExists = errors.New("given static key already exists")
	ErrURLNotFound      = urlrepo.ErrURLNotFound
)

type URLSvc interface {
	Create(ctx context.Context, address string, expire *time.Time) (string, error)
	CreateWithKey(ctx context.Context, key string, address string, expire *time.Time) error
	Visit(ctx context.Context, key string) (url.URL, error)
}

type urlSvc struct {
	repo   urlrepo.Repository
	logger *zap.Logger
	gen    generator.Generator
}

func ProvideURLSvc(repo urlrepo.Repository, logger *zap.Logger, gen generator.Generator) URLSvc {
	return &urlSvc{
		gen:    gen,
		repo:   repo,
		logger: logger.Named("urlsvc"),
	}
}

// MaxKeyGenAttempts bounds how many keys Create will try before giving up. Generators that
// build unique keys by construction never collide, and for random ones the expected number of
// attempts is 1/(1-load factor) — so this is only ever reached once the key space is nearly full.
const MaxKeyGenAttempts = 5

func (s *urlSvc) Create(ctx context.Context, address string, expire *time.Time) (string, error) {
	for attempt := range MaxKeyGenAttempts {
		key, err := s.gen.ShortURLKey(ctx)
		if err != nil {
			return "", fmt.Errorf("short url key generation failed %w", err)
		}

		err = s.create(ctx, key, address, expire)
		if err == nil {
			return key, nil
		}

		if !errors.Is(err, urlrepo.ErrDuplicateShortURL) {
			return "", err
		}

		s.logger.Warn("short url key already taken, generating another one",
			zap.String("key", key), zap.Int("attempt", attempt+1))
	}

	return "", ErrKeyGenFailed
}

func (s *urlSvc) CreateWithKey(ctx context.Context, key string, address string, expire *time.Time) error {
	key = "static_" + key

	if err := s.create(ctx, key, address, expire); err != nil {
		if errors.Is(err, urlrepo.ErrDuplicateShortURL) {
			return ErrKeyAlreadyExists
		}

		return err
	}

	return nil
}

func (s *urlSvc) Visit(ctx context.Context, key string) (url.URL, error) {
	url, err := s.visit(ctx, key)
	if err != nil {
		return url, err
	}

	// use atomic increment to avoid race conditions
	if err := s.repo.IncrementVisits(ctx, url.Key); err != nil {
		s.logger.Error("incrementing url visit count failed", zap.Error(err))
	}

	url.Visits++

	return url, nil
}

func (s *urlSvc) visit(ctx context.Context, key string) (url.URL, error) {
	{
		url, err := s.repo.FromShortURL(ctx, key)
		if err != nil {
			if !errors.Is(err, urlrepo.ErrURLNotFound) {
				return url, fmt.Errorf("url fetching failed %w", err)
			}
		} else {
			return url, nil
		}
	}

	url, err := s.repo.FromShortURL(ctx, "static_"+key)
	if err != nil {
		if errors.Is(err, urlrepo.ErrURLNotFound) {
			return url, ErrURLNotFound
		}

		return url, fmt.Errorf("url fetching failed %w", err)
	}

	return url, nil
}

func (s *urlSvc) create(ctx context.Context, key string, address string, expire *time.Time) error {
	valid := true

	if expire == nil {
		expire = new(time.Time)
		valid = false
	}

	// nolint exhaustruct
	url := url.URL{
		Key:    key,
		URL:    address,
		Visits: 0,
		Expire: sql.NullTime{
			Time:  *expire,
			Valid: valid,
		},
	}

	if err := s.repo.Create(ctx, url); err != nil {
		return fmt.Errorf("url creation failed %w", err)
	}

	return nil
}
