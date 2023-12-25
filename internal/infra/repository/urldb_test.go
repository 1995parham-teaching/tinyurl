package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1995parham-teaching/tinyurl/internal/infra/config"
	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
	"github.com/1995parham-teaching/tinyurl/internal/infra/logger"
	"github.com/1995parham-teaching/tinyurl/internal/infra/repository"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"
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

func (s *URLDBTestSuite) TearDownTest() {
	require := s.Require()

	// nolint: exhaustruct
	stmt := &gorm.Statement{DB: s.db.DB}
	require.NoError(stmt.Parse(new(url.URL)))

	tx := s.db.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s;", stmt.Schema.Table))
	require.NoError(tx.Error)
}

func (s *URLDBTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *URLDBTestSuite) TestNotFound() {
	require := s.Require()

	_, err := s.repo.FromShortURL(context.Background(), "static_random")
	require.ErrorIs(urlrepo.ErrURLNotFound, err)
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
	require.False(url.Expire.Valid)
}

func TestURLDB(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(URLDBTestSuite))
}
