package generator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/stretchr/testify/require"
)

// errSequence stands in for a database that will not hand out identifiers.
var errSequence = errors.New("sequence is unavailable")

// fakeSequencer replaces the postgres sequence so the counting generators can be exercised
// without a database.
type fakeSequencer struct {
	next uint64
	err  error
}

func (f *fakeSequencer) NextID(_ context.Context) (uint64, error) {
	if f.err != nil {
		return 0, f.err
	}

	id := f.next
	f.next++

	return id, nil
}

func TestCounter(t *testing.T) {
	t.Parallel()

	c := generator.NewCounter(&fakeSequencer{next: 0, err: nil})
	require.Implements(t, new(generator.Generator), c)

	key, err := c.ShortURLKey(t.Context())
	require.NoError(t, err)
	requireWellFormed(t, key)
	require.Equal(t, "000000", key)
}

// TestCounterNeverRepeats is the whole point of the strategy: no collision check is performed
// anywhere, so the encoding of a sequence must be enough on its own.
func TestCounterNeverRepeats(t *testing.T) {
	t.Parallel()

	const draws = 20000

	c := generator.NewCounter(&fakeSequencer{next: 0, err: nil})
	keys := make(map[string]struct{}, draws)

	for range draws {
		key, err := c.ShortURLKey(t.Context())
		require.NoError(t, err)
		requireWellFormed(t, key)

		_, duplicate := keys[key]
		require.False(t, duplicate, "counter handed out %q twice", key)

		keys[key] = struct{}{}
	}
}

func TestCounterSequenceFailure(t *testing.T) {
	t.Parallel()

	c := generator.NewCounter(&fakeSequencer{next: 0, err: errSequence})

	_, err := c.ShortURLKey(t.Context())
	require.ErrorIs(t, err, errSequence)
}

func TestCounterBeyondKeySpace(t *testing.T) {
	t.Parallel()

	c := generator.NewCounter(&fakeSequencer{next: generator.Space, err: nil})

	_, err := c.ShortURLKey(t.Context())
	require.ErrorIs(t, err, generator.ErrKeySpaceExhausted)
}
