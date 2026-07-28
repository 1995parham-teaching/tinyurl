package generator_test

import (
	"testing"

	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/stretchr/testify/require"
)

// TestProvide covers the types that need no database. The counting ones are built against a
// real sequence and are exercised in their own tests through a fake.
func TestProvide(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		cfg  generator.Config
	}{
		{name: "simple", cfg: generator.Config{Type: generator.TypeSimple, Key: ""}},
		{name: "secure", cfg: generator.Config{Type: generator.TypeSecure, Key: ""}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gen, err := generator.Provide(tc.cfg, nil)
			require.NoError(t, err)
			require.NotNil(t, gen)

			key, err := gen.ShortURLKey(t.Context())
			require.NoError(t, err)
			requireWellFormed(t, key)
		})
	}
}

// TestProvideUnknownType pins the decision to fail loudly: a typo in the configuration used to
// fall through to the simple generator, silently running a strategy nobody asked for.
func TestProvideUnknownType(t *testing.T) {
	t.Parallel()

	_, err := generator.Provide(generator.Config{Type: "typo", Key: ""}, nil)
	require.ErrorIs(t, err, generator.ErrUnknownType)
}

func TestProvideFeistelWithoutKey(t *testing.T) {
	t.Parallel()

	_, err := generator.Provide(generator.Config{Type: generator.TypeFeistel, Key: ""}, nil)
	require.ErrorIs(t, err, generator.ErrMissingFeistelKey)
}
