package generator_test

import (
	"testing"

	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/stretchr/testify/require"
)

func TestSecure(t *testing.T) {
	t.Parallel()

	s := new(generator.Secure)
	require.Implements(t, new(generator.Generator), s)

	key, err := s.ShortURLKey(t.Context())
	require.NoError(t, err)
	requireWellFormed(t, key)
}

// TestSecureCoversAlphabet checks that rejection sampling has not accidentally excluded the
// digits at the end of the alphabet, which is exactly what a naive byte modulo would do.
func TestSecureCoversAlphabet(t *testing.T) {
	t.Parallel()

	const draws = 5000

	s := new(generator.Secure)
	seen := make(map[rune]struct{}, generator.Base)

	for range draws {
		key, err := s.ShortURLKey(t.Context())
		require.NoError(t, err)
		requireWellFormed(t, key)

		for _, digit := range key {
			seen[digit] = struct{}{}
		}
	}

	require.Len(t, seen, generator.Base, "some digits of the alphabet never came up")
}
