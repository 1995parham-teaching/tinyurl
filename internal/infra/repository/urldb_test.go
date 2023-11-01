package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

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

	repo urlrepo.Repository
	db   *db.DB

	app *fxtest.App
}

func (s *URLDBTestSuite) SetupSuite() {
	s.app = fxtest.New(s.T(),
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(db.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(
			fx.Annotate(repository.ProvideURLDB, fx.As(new(urlrepo.Repository))),
		),
		fx.Invoke(func(repo urlrepo.Repository, db *db.DB) {
			s.db = db
			s.repo = repo
		}),
	).RequireStart()
}

func (s *URLDBTestSuite) TearDownTest() {}

func (s *URLDBTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *URLDBTestSuite) TestCreate() {
	require := s.Require()

	// nolint: exhaustruct
	require.NoError(s.repo.Create(context.Background(), url.URL{
		Key:    "static_random",
		URL:    "https://github.com",
		Visits: 0,
		Expire: sql.NullTime{
			Time:  time.Now(),
			Valid: false,
		},
	}))

	// nolint: exhaustruct
	url, err := s.repo.FromShortURL(context.Background(), "static_random")
	require.NoError(err)

	require.Equal("https://github.com", url.URL)
	require.NoError(s.db.DB.Where("key = ?", "static_random").Delete(&url).Error)
}

func TestURLDB(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(URLDBTestSuite))
}
