package generator

import (
	"context"
	"errors"
	"fmt"

	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
)

// Supported generator types. They differ in where uniqueness comes from: the random ones lean
// on the primary key of urls and a retry, the counting ones are unique by construction.
const (
	// TypeSimple draws random keys from math/rand.
	TypeSimple = "simple"
	// TypeSecure draws random keys from crypto/rand.
	TypeSecure = "secure"
	// TypeCounter encodes a database sequence, yielding sequential keys.
	TypeCounter = "counter"
	// TypeFeistel scrambles a database sequence, yielding keys that are unique but unordered.
	TypeFeistel = "feistel"
)

// ErrUnknownType is returned for a generator type that is not implemented.
var ErrUnknownType = errors.New("unknown generator type")

type Generator interface {
	ShortURLKey(ctx context.Context) (string, error)
}

// Provide returns a Generator based on the config type. An unknown type is an error rather
// than a silent fallback, so that a typo fails at startup instead of quietly changing which
// strategy the service runs.
func Provide(cfg Config, database *db.DB) (Generator, error) {
	switch cfg.Type {
	case TypeSimple:
		return new(Simple), nil
	case TypeSecure:
		return new(Secure), nil
	case TypeCounter:
		return NewCounter(NewPostgresSequence(database)), nil
	case TypeFeistel:
		gen, err := NewFeistel(NewPostgresSequence(database), cfg.Key)
		if err != nil {
			return nil, fmt.Errorf("building the feistel generator failed %w", err)
		}

		return gen, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownType, cfg.Type)
	}
}
