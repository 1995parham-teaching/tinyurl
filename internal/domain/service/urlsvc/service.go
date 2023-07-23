package urlsvc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
	"github.com/1989michael/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1989michael/tinyurl/internal/infra/generator"
	"go.uber.org/zap"
)

var (
	ErrKeyGenFailed     = errors.New("cannot generate new random string as short url")
	ErrKeyAlreadyExists = errors.New("given static key already exists")
	ErrURLNotFound      = urlrepo.ErrURLNotFound
)

type URLSvc struct {
	repo   urlrepo.Repository
	logger *zap.Logger
	gen    generator.Generator
}

func ProvideURLSvc(repo urlrepo.Repository, logger *zap.Logger, gen generator.Generator) *URLSvc {
	return &URLSvc{
		gen:    gen,
		repo:   repo,
		logger: logger.Named("urlsvc"),
	}
}

func (s *URLSvc) create(ctx context.Context, key string, address string, expire *time.Time) error {
	// nolint exhaustruct
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
	key := s.gen.ShortURLKey()

	if err := s.create(ctx, key, address, expire); err != nil {
		if errors.Is(err, urlrepo.ErrDuplicateShortURL) {
			return "", ErrKeyGenFailed
		}

		return "", err
	}

	return key, nil
}

func (s *URLSvc) CreateWithKey(ctx context.Context, key string, address string, expire *time.Time) error {
	key = fmt.Sprintf("static_%s", key)

	if err := s.create(ctx, key, address, expire); err != nil {
		if errors.Is(err, urlrepo.ErrDuplicateShortURL) {
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
			if !errors.Is(err, urlrepo.ErrURLNotFound) {
				return url, fmt.Errorf("url fetching failed %w", err)
			}
		} else {
			return url, nil
		}
	}

	url, err := s.repo.FromShortURL(ctx, fmt.Sprintf("static_%s", key))
	if err != nil {
		if errors.Is(err, urlrepo.ErrURLNotFound) {
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

	// we can use transaction here but number of visits is not accurate number.
	url.Visits++

	if err := s.repo.Update(ctx, url); err != nil {
		s.logger.Error("updating url visit coount failed", zap.Error(err))
	}

	return url, nil
}
