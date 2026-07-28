package generator_test

import (
	"strings"
	"testing"

	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/stretchr/testify/require"
)

// requireWellFormed asserts that a key is something the service can hand out and later look up.
func requireWellFormed(t *testing.T, key string) {
	t.Helper()

	require.Len(t, key, generator.KeyLength)

	for _, digit := range key {
		require.True(t, strings.ContainsRune(generator.Alphabet, digit),
			"key %q contains %q which is outside the alphabet", key, digit)
	}
}

func TestSimple(t *testing.T) {
	t.Parallel()

	s := new(generator.Simple)
	require.Implements(t, new(generator.Generator), s)

	key, err := s.ShortURLKey(t.Context())
	require.NoError(t, err)
	requireWellFormed(t, key)
}

// TestSimpleSpreads is a smoke test that the generator is drawing from the whole alphabet
// rather than repeating itself. It is not a uniqueness guarantee — Simple has none, the
// primary key on urls does.
func TestSimpleSpreads(t *testing.T) {
	t.Parallel()

	const draws = 1000

	s := new(generator.Simple)
	keys := make(map[string]struct{}, draws)

	for range draws {
		key, err := s.ShortURLKey(t.Context())
		require.NoError(t, err)
		requireWellFormed(t, key)

		keys[key] = struct{}{}
	}

	require.Len(t, keys, draws, "random keys repeated within %d draws", draws)
}
