package generator_test

import (
	"testing"

	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/stretchr/testify/require"
)

const feistelKey = "test-secret"

func newFeistel(t *testing.T, seq generator.Sequencer) *generator.Feistel {
	t.Helper()

	f, err := generator.NewFeistel(seq, feistelKey)
	require.NoError(t, err)

	return f
}

func TestFeistel(t *testing.T) {
	t.Parallel()

	f := newFeistel(t, &fakeSequencer{next: 0, err: nil})
	require.Implements(t, new(generator.Generator), f)

	key, err := f.ShortURLKey(t.Context())
	require.NoError(t, err)
	requireWellFormed(t, key)
}

func TestFeistelRequiresKey(t *testing.T) {
	t.Parallel()

	_, err := generator.NewFeistel(&fakeSequencer{next: 0, err: nil}, "")
	require.ErrorIs(t, err, generator.ErrMissingFeistelKey)
}

// TestFeistelIsBijective is the property that lets the generator skip collision checks: the
// permutation must map distinct identifiers to distinct results, all of them inside the key
// space so that they can actually be encoded.
func TestFeistelIsBijective(t *testing.T) {
	t.Parallel()

	const ids = 50000

	f := newFeistel(t, &fakeSequencer{next: 0, err: nil})
	seen := make(map[uint64]uint64, ids)

	for id := range uint64(ids) {
		permuted := f.Permute(id)

		require.Less(t, permuted, generator.Space, "identifier %d was permuted outside the key space", id)

		previous, duplicate := seen[permuted]
		require.False(t, duplicate, "identifiers %d and %d both permuted to %d", previous, id, permuted)

		seen[permuted] = id
	}
}

// TestFeistelIsDeterministic matters because the same identifier must never be encoded two
// different ways, and because a restart must not start reissuing keys already handed out.
func TestFeistelIsDeterministic(t *testing.T) {
	t.Parallel()

	first := newFeistel(t, &fakeSequencer{next: 0, err: nil})
	second := newFeistel(t, &fakeSequencer{next: 0, err: nil})

	for id := range uint64(100) {
		require.Equal(t, first.Permute(id), second.Permute(id))
	}
}

// TestFeistelDependsOnKey checks that the secret really drives the permutation, which is what
// stops an observer from predicting the next key.
func TestFeistelDependsOnKey(t *testing.T) {
	t.Parallel()

	mine := newFeistel(t, &fakeSequencer{next: 0, err: nil})

	theirs, err := generator.NewFeistel(&fakeSequencer{next: 0, err: nil}, "another-secret")
	require.NoError(t, err)

	differences := 0

	for id := range uint64(100) {
		if mine.Permute(id) != theirs.Permute(id) {
			differences++
		}
	}

	require.Greater(t, differences, 90, "changing the secret barely changed the permutation")
}

// TestFeistelScatters is what separates this generator from Counter: consecutive identifiers
// must not produce keys an observer can walk from one to the next.
func TestFeistelScatters(t *testing.T) {
	t.Parallel()

	const (
		ids           = 1000
		neighbourhood = 1000
	)

	f := newFeistel(t, &fakeSequencer{next: 0, err: nil})

	adjacent := 0

	for id := range uint64(ids) {
		current, next := f.Permute(id), f.Permute(id+1)

		if current < next && next-current < neighbourhood {
			adjacent++
		} else if next < current && current-next < neighbourhood {
			adjacent++
		}
	}

	require.Less(t, adjacent, ids/10, "consecutive identifiers landed close together too often")
}

func TestFeistelSequenceFailure(t *testing.T) {
	t.Parallel()

	f := newFeistel(t, &fakeSequencer{next: 0, err: errSequence})

	_, err := f.ShortURLKey(t.Context())
	require.ErrorIs(t, err, errSequence)
}

func TestFeistelBeyondKeySpace(t *testing.T) {
	t.Parallel()

	f := newFeistel(t, &fakeSequencer{next: generator.Space, err: nil})

	_, err := f.ShortURLKey(t.Context())
	require.ErrorIs(t, err, generator.ErrKeySpaceExhausted)
}
