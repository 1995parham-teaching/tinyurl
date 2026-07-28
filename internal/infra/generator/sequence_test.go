package generator_test

import (
	"context"
	"testing"

	"github.com/1995parham-teaching/tinyurl/internal/infra/config"
	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/1995parham-teaching/tinyurl/internal/infra/logger"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type PostgresSequenceTestSuite struct {
	suite.Suite

	seq *generator.PostgresSequence

	app *fxtest.App
}

func (s *PostgresSequenceTestSuite) SetupSuite() {
	s.app = fxtest.New(s.T(),
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(db.Provide),
		fx.Provide(generator.NewPostgresSequence),
		fx.Invoke(func(seq *generator.PostgresSequence) {
			s.seq = seq
		}),
	).RequireStart()
}

func (s *PostgresSequenceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

// TestNextIDIncreases pins the contract the counting generators depend on: identifiers are
// never handed out twice. Gaps are expected, because the sequence is cached per session.
func (s *PostgresSequenceTestSuite) TestNextIDIncreases() {
	require := s.Require()

	previous, err := s.seq.NextID(context.Background())
	require.NoError(err)

	for range 100 {
		current, err := s.seq.NextID(context.Background())
		require.NoError(err)
		require.Greater(current, previous)

		previous = current
	}
}

// TestGeneratorsRun exercises the counting generators against the real sequence, which is the
// only place the nextval query itself is covered.
func (s *PostgresSequenceTestSuite) TestGeneratorsRun() {
	require := s.Require()

	feistel, err := generator.NewFeistel(s.seq, feistelKey)
	require.NoError(err)

	generators := map[string]generator.Generator{
		generator.TypeCounter: generator.NewCounter(s.seq),
		generator.TypeFeistel: feistel,
	}

	for name, gen := range generators {
		key, err := gen.ShortURLKey(context.Background())
		require.NoError(err, "%s generator failed", name)
		require.Len(key, generator.KeyLength)
	}
}

func TestPostgresSequence(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(PostgresSequenceTestSuite))
}
