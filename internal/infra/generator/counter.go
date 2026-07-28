package generator

import (
	"context"
	"fmt"
)

// Counter turns a monotonically increasing identifier into a key.
//
// Uniqueness is structural: distinct identifiers encode to distinct keys, so there is no
// collision to detect, nothing to retry, and no wasted round trip asking the database whether
// a key is free. It also packs the key space densely instead of scattering keys across it.
//
// The catch is that keys are sequential, so anyone holding one can walk the whole service by
// counting. Feistel keeps this generator's guarantee and removes that property.
type Counter struct {
	seq Sequencer
}

func NewCounter(seq Sequencer) *Counter {
	return &Counter{seq: seq}
}

// ShortURLKey encodes the next identifier of the sequence.
func (c *Counter) ShortURLKey(ctx context.Context) (string, error) {
	id, err := c.seq.NextID(ctx)
	if err != nil {
		return "", fmt.Errorf("drawing an identifier failed %w", err)
	}

	key, err := Encode(id)
	if err != nil {
		return "", fmt.Errorf("encoding identifier %d failed %w", id, err)
	}

	return key, nil
}
