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

	// nolint: exhaustruct_v5
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

	_, err := s.repo.FromShortURL(context.Background(), "random")
	require.ErrorIs(urlrepo.ErrURLNotFound, err)
}

func (s *URLDBTestSuite) TestCreate() {
	require := s.Require()

	// nolint: exhaustruct_v5
	require.NoError(s.repo.Create(context.Background(), url.URL{
		Key:    "random",
		URL:    testURL,
		Visits: 0,
		Expire: sql.NullTime{
			Time:  time.Now(),
			Valid: false,
		},
	}))

	// nolint: exhaustruct_v5
	url, err := s.repo.FromShortURL(context.Background(), "random")
	require.NoError(err)

	require.Equal(testURL, url.URL)
	require.False(url.Expire.Valid)
}

// TestDuplicateKey covers the guarantee the whole scheme rests on: the primary key is what
// makes short urls unique, and the repository has to report a violation as such so the service
// can generate another key instead of failing the request.
func (s *URLDBTestSuite) TestDuplicateKey() {
	require := s.Require()

	// nolint: exhaustruct_v5
	record := url.URL{
		Key:    "taken",
		URL:    testURL,
		Visits: 0,
		Expire: sql.NullTime{Time: time.Now(), Valid: false},
	}

	require.NoError(s.repo.Create(context.Background(), record))

	record.URL = "https://gitlab.com"
	require.ErrorIs(s.repo.Create(context.Background(), record), urlrepo.ErrDuplicateShortURL)
}

// TestExpired checks that an expired url is indistinguishable from a missing one. The row is
// backdated because a CHECK constraint refuses an expiry that precedes creation.
func (s *URLDBTestSuite) TestExpired() {
	require := s.Require()

	created := time.Now().Add(-2 * time.Hour)

	// nolint: exhaustruct_v5
	require.NoError(s.repo.Create(context.Background(), url.URL{
		Key:    "expired",
		URL:    testURL,
		Visits: 0,
		Expire: sql.NullTime{
			Time:  created.Add(time.Hour),
			Valid: true,
		},
		CreatedAt: created,
		UpdatedAt: created,
	}))

	_, err := s.repo.FromShortURL(context.Background(), "expired")
	require.ErrorIs(err, urlrepo.ErrURLNotFound)
}

func (s *URLDBTestSuite) TestNotExpired() {
	require := s.Require()

	// nolint: exhaustruct_v5
	require.NoError(s.repo.Create(context.Background(), url.URL{
		Key:    "fresh",
		URL:    testURL,
		Visits: 0,
		Expire: sql.NullTime{
			Time:  time.Now().Add(time.Hour),
			Valid: true,
		},
	}))

	record, err := s.repo.FromShortURL(context.Background(), "fresh")
	require.NoError(err)
	require.Equal(testURL, record.URL)
}

// TestIncrementVisitsBatch covers the statement that folds a whole batch of counters into one
// round trip, which is the only place the VALUES join is exercised against a real database.
func (s *URLDBTestSuite) TestIncrementVisitsBatch() {
	require := s.Require()

	const (
		first  = "one"
		second = "two"
		// a key that no longer exists must not fail the batch, since a url may be removed
		// between a visit being counted and the batch reaching the database.
		removed = "gone"
	)

	for _, key := range []string{first, second} {
		// nolint: exhaustruct_v5
		require.NoError(s.repo.Create(context.Background(), url.URL{
			Key:    key,
			URL:    testURL,
			Visits: 0,
			Expire: sql.NullTime{Time: time.Now(), Valid: false},
		}))
	}

	require.NoError(s.repo.IncrementVisitsBatch(context.Background(), map[string]uint64{
		first:   7,
		second:  3,
		removed: 5,
	}))

	one, err := s.repo.FromShortURL(context.Background(), first)
	require.NoError(err)
	require.Equal(uint64(7), one.Visits)

	two, err := s.repo.FromShortURL(context.Background(), second)
	require.NoError(err)
	require.Equal(uint64(3), two.Visits)

	// batches accumulate rather than overwrite.
	require.NoError(s.repo.IncrementVisitsBatch(context.Background(), map[string]uint64{first: 2}))

	one, err = s.repo.FromShortURL(context.Background(), first)
	require.NoError(err)
	require.Equal(uint64(9), one.Visits)
}

func (s *URLDBTestSuite) TestIncrementVisitsBatchEmpty() {
	s.Require().NoError(s.repo.IncrementVisitsBatch(context.Background(), map[string]uint64{}))
}

func TestURLDB(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(URLDBTestSuite))
}
