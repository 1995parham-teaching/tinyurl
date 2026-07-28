package generator

import (
	"context"
	"math/rand/v2"
)

// Simple is an easy to use random key generator.
//
// It offers no uniqueness guarantee of its own: the primary key on urls.key rejects a
// duplicate and the caller generates another one. That works because the key space stays
// sparse — with N keys stored, a fresh key collides with probability N/Space, and the expected
// number of attempts is 1/(1-load factor). Keys are drawn from math/rand and so are
// predictable; use Secure where a third party must not be able to guess them.
type Simple struct{}

// ShortURLKey generates a random key from the source characters.
func (Simple) ShortURLKey(_ context.Context) (string, error) {
	key := make([]byte, KeyLength)
	for i := range key {
		key[i] = Alphabet[rand.IntN(Base)] // nolint: gosec
	}

	return string(key), nil
}
