package generator_test

import (
	"testing"

	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		id       uint64
		expected string
	}{
		{name: "zero", id: 0, expected: "000000"},
		{name: "one", id: 1, expected: "000001"},
		{name: "last digit", id: generator.Base - 1, expected: "00000z"},
		{name: "carry", id: generator.Base, expected: "000010"},
		{name: "last key", id: generator.Space - 1, expected: "zzzzzz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key, err := generator.Encode(tc.id)
			require.NoError(t, err)
			require.Equal(t, tc.expected, key)
		})
	}
}

func TestEncodeBeyondKeySpace(t *testing.T) {
	t.Parallel()

	_, err := generator.Encode(generator.Space)
	require.ErrorIs(t, err, generator.ErrKeySpaceExhausted)
}

// TestEncodeIsInjective is the property the counting generators rest on: distinct identifiers
// must never encode to the same key, otherwise uniqueness by construction does not hold.
func TestEncodeIsInjective(t *testing.T) {
	t.Parallel()

	const ids = 20000

	keys := make(map[string]struct{}, ids)

	for id := range uint64(ids) {
		key, err := generator.Encode(id)
		require.NoError(t, err)

		_, duplicate := keys[key]
		require.False(t, duplicate, "identifier %d encoded to an already used key %q", id, key)

		keys[key] = struct{}{}
	}
}
