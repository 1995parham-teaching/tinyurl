package generator

import (
	"context"
	"fmt"

	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
)

// SequenceName is the postgres sequence backing the counting generators.
const SequenceName = "url_id_seq"

// nextValueQuery draws the next identifier. The sequence name is a compile time constant, so
// there is nothing here for a caller to inject.
const nextValueQuery = `SELECT nextval('` + SequenceName + `')`

// Sequencer hands out identifiers that are never repeated. Generators built on top of one are
// collision free by construction: uniqueness comes from the counter, not from asking the
// database whether a key is taken and retrying when it is.
type Sequencer interface {
	NextID(ctx context.Context) (uint64, error)
}

// PostgresSequence draws identifiers from a postgres sequence.
//
// The sequence is declared with a cache, so each database session claims a block of values up
// front and serves them without further coordination. That is the same idea as a hand written
// Hi/Lo allocator, minus the code. The price is gaps in the numbering whenever a session ends
// with values unused, which costs nothing but key space.
type PostgresSequence struct {
	db *db.DB
}

func NewPostgresSequence(database *db.DB) *PostgresSequence {
	return &PostgresSequence{db: database}
}

func (s *PostgresSequence) NextID(ctx context.Context) (uint64, error) {
	var id uint64

	if err := s.db.DB.WithContext(ctx).Raw(nextValueQuery).Scan(&id).Error; err != nil {
		return 0, fmt.Errorf("fetching next value of sequence %s failed %w", SequenceName, err)
	}

	return id, nil
}
