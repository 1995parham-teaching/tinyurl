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

func (s *urlSvc) Create(ctx context.Context, address string, expire *time.Time) (string, error) {
	key := s.gen.ShortURLKey()

	if err := s.create(ctx, key, address, expire); err != nil {
		if errors.Is(err, urlrepo.ErrDuplicateShortURL) {
			return "", ErrKeyGenFailed
		}

		return "", err
	}

	return key, nil
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

	// we can use transaction here but number of visits is not accurate number.
	url.Visits++

	if err := s.repo.Update(ctx, url); err != nil {
		s.logger.Error("updating url visit coount failed", zap.Error(err))
	}

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
