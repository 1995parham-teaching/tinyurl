package generator

import (
	"context"
	"crypto/rand"
	"fmt"
)

// byteValues is the number of distinct values a byte can hold.
const byteValues = 256

// unbiasedLimit is the largest byte value that survives reduction modulo Base without skewing
// the result. 256 is not a multiple of 62, so the digits at the start of the alphabet would
// otherwise come up more often than the ones at the end.
const unbiasedLimit = byteValues - byteValues%Base

// Secure is Simple with an unpredictable source.
//
// Uniqueness works exactly as it does for Simple — the primary key rejects duplicates and the
// caller retries. What changes is that keys cannot be guessed or enumerated, which matters
// because knowing a key is the only thing standing between a stranger and the target url.
type Secure struct{}

// ShortURLKey generates a random key using crypto/rand, discarding the byte values that
// would bias the choice of digit.
func (Secure) ShortURLKey(_ context.Context) (string, error) {
	key := make([]byte, 0, KeyLength)
	buf := make([]byte, KeyLength)

	for len(key) < KeyLength {
		if _, err := rand.Read(buf); err != nil {
			return "", fmt.Errorf("reading random bytes failed %w", err)
		}

		for _, b := range buf {
			if b >= unbiasedLimit {
				continue
			}

			key = append(key, Alphabet[b%Base])

			if len(key) == KeyLength {
				break
			}
		}
	}

	return string(key), nil
}
