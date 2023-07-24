package repository_test

import (
	"context"
	"testing"

	"github.com/1989michael/tinyurl/internal/domain/model/url"
	"github.com/1989michael/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1989michael/tinyurl/internal/infra/config"
	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/1989michael/tinyurl/internal/infra/logger"
	"github.com/1989michael/tinyurl/internal/infra/repository"
	"github.com/1989michael/tinyurl/internal/infra/telemetry"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type URLDBTestSuite struct {
	suite.Suite
	options []fx.Option
}

func (s *URLDBTestSuite) Invoke(f any) {
	options := []fx.Option{
		fx.Invoke(f),
	}
	options = append(options, s.options...)

	fxtest.New(s.T(), options...).RequireStart().RequireStop()
}

func (s *URLDBTestSuite) SetupSuite() {
	s.options = []fx.Option{
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(db.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(
			fx.Annotate(repository.ProvideURLDB, fx.As(new(urlrepo.Repository))),
		),
	}
}

func (s *URLDBTestSuite) TestCreate() {
	s.Invoke(s.testCreate)
}

func (s *URLDBTestSuite) testCreate(repo urlrepo.Repository, db *db.DB) {
	require := s.Require()

	// nolint: exhaustruct
	require.NoError(repo.Create(context.Background(), url.URL{
		Key:    "static_random",
		URL:    "https://github.com",
		Visits: 0,
		Expire: nil,
	}))

	// nolint: exhaustruct
	url, err := repo.FromShortURL(context.Background(), "static_random")
	require.NoError(err)

	require.Equal(url.URL, "https://github.com")

	require.NoError(db.DB.Where("key = ?", "static_random").Delete(&url).Error)
}

func TestURLDB(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(URLDBTestSuite))
}
