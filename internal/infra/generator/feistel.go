package generator

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// feistelHalfBits splits the block in two equal halves. Two 18 bit halves make a 36 bit
	// block, and 2**36 is the smallest power of two that covers Space.
	feistelHalfBits = 18
	feistelHalfMask = 1<<feistelHalfBits - 1

	// feistelRounds is how many times the halves are mixed. Four is the usual floor for a
	// Feistel network to look random rather than merely shuffled.
	feistelRounds = 4
)

// ErrMissingFeistelKey is returned when the feistel generator is configured without a secret.
var ErrMissingFeistelKey = errors.New("feistel generator requires a secret key")

// Feistel keeps Counter's guarantee and drops its predictability.
//
// It draws a sequential identifier and pushes it through a keyed Feistel network before
// encoding it. A Feistel network is a bijection whatever its round function does, so distinct
// identifiers still map to distinct keys: uniqueness remains structural, with no collision
// check and no retry. What the network buys is that consecutive identifiers land far apart, so
// holding one key tells an observer nothing about any other.
//
// The secret key is what makes the permutation unguessable. Changing it re-shuffles every
// identifier that has not been handed out yet; keys already stored keep working, since lookup
// never inverts the permutation.
type Feistel struct {
	seq Sequencer
	key []byte
}

func NewFeistel(seq Sequencer, key string) (*Feistel, error) {
	if key == "" {
		return nil, ErrMissingFeistelKey
	}

	return &Feistel{seq: seq, key: []byte(key)}, nil
}

// ShortURLKey encodes the next identifier of the sequence after scrambling it.
func (f *Feistel) ShortURLKey(ctx context.Context) (string, error) {
	id, err := f.seq.NextID(ctx)
	if err != nil {
		return "", fmt.Errorf("drawing an identifier failed %w", err)
	}

	if id >= Space {
		return "", fmt.Errorf("%w: identifier %d has no place to be permuted into", ErrKeySpaceExhausted, id)
	}

	key, err := Encode(f.Permute(id))
	if err != nil {
		return "", fmt.Errorf("encoding identifier %d failed %w", id, err)
	}

	return key, nil
}

// Permute maps [0, Space) onto itself bijectively.
//
// The underlying network permutes the whole 36 bit block, which is larger than Space, so a
// result landing outside the key space is fed back through until it lands inside. That is
// cycle walking, and it preserves bijectivity: following a permutation forward from a point of
// the subset always arrives at another point of the subset, and no two starting points arrive
// at the same one. It terminates because a permutation of a finite set is made of cycles, so
// walking forward from a member of the subset must eventually return to it.
func (f *Feistel) Permute(id uint64) uint64 {
	for {
		id = f.round(id)

		if id < Space {
			return id
		}
	}
}

// round runs the block through the Feistel network once.
func (f *Feistel) round(block uint64) uint64 {
	left, right := block>>feistelHalfBits&feistelHalfMask, block&feistelHalfMask

	for i := range byte(feistelRounds) {
		left, right = right, left^f.mix(i, right)
	}

	return left<<feistelHalfBits | right
}

// mix is the round function. It may be any function at all without endangering the bijection,
// which is the whole appeal of the construction; a keyed MAC is what makes the result look
// random to somebody who does not hold the key.
func (f *Feistel) mix(round byte, half uint64) uint64 {
	var buf [9]byte

	buf[0] = round
	binary.BigEndian.PutUint64(buf[1:], half)

	mac := hmac.New(sha256.New, f.key)
	// hash.Hash documents that Write never returns an error.
	_, _ = mac.Write(buf[:])

	return binary.BigEndian.Uint64(mac.Sum(nil)) & feistelHalfMask
}
