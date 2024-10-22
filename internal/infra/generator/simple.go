package generator

import "math/rand/v2"

// Simple is an easy to use random key generator.
type Simple struct{}

// ShortURLKey generates a random key from the source characters.
func (Simple) ShortURLKey() string {
	const (
		length = 6
		source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	)

	b := make([]byte, length)
	for i := range b {
		b[i] = source[rand.IntN(len(source))] // nolint: gosec
	}

	return string(b)
}
