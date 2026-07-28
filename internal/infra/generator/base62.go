package generator

import (
	"errors"
	"fmt"
)

const (
	// Alphabet holds the digits a short url key is written with. Every generator draws from it
	// so that keys are interchangeable no matter which strategy produced them.
	Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	// Base is the radix of Alphabet.
	Base = 62

	// KeyLength is the number of characters in a generated short url key.
	KeyLength = 6

	// Space is the number of distinct keys, Base ** KeyLength. Random generators care about
	// how full it gets (collisions rise with the load factor) and counting generators care
	// about not running past the end of it.
	Space = uint64(Base) * Base * Base * Base * Base * Base
)

// compile time assertion that Alphabet really holds Base digits: whichever way the two drift
// apart, one of these differences goes negative and a negative constant has no uint form.
const (
	_ = uint(len(Alphabet) - Base)
	_ = uint(Base - len(Alphabet))
)

// ErrKeySpaceExhausted is returned when an identifier no longer fits in KeyLength characters.
var ErrKeySpaceExhausted = errors.New("short url key space is exhausted")

// Encode writes id as a KeyLength character base62 key, left padded with the zero digit.
// It is injective: distinct ids below Space always encode to distinct keys.
func Encode(id uint64) (string, error) {
	if id >= Space {
		return "", fmt.Errorf("%w: %d does not fit in %d characters", ErrKeySpaceExhausted, id, KeyLength)
	}

	key := make([]byte, KeyLength)

	for i := KeyLength - 1; i >= 0; i-- {
		key[i] = Alphabet[id%Base]
		id /= Base
	}

	return string(key), nil
}
