package repository

import "time"

// Config tunes the layers that sit between the service and the database.
type Config struct {
	Cache  CacheConfig  `json:"cache"  koanf:"cache"`
	Visits VisitsConfig `json:"visits" koanf:"visits"`
}

// CacheConfig tunes the read path.
//
// A shortener serves far more redirects than it accepts creations, and what a redirect needs —
// the target of a key — never changes once written. That makes it about the most cacheable
// thing there is.
type CacheConfig struct {
	// Enabled turns the cache off entirely, which is useful for seeing what it buys.
	Enabled bool `json:"enabled" koanf:"enabled"`

	// Size is how many entries to keep before evicting the least recently used one.
	Size int `json:"size" koanf:"size"`

	// TTL bounds how long an entry is trusted. Entries are also capped by the expiry of the
	// url they hold, so a url can never outlive its expiry by sitting in the cache.
	TTL time.Duration `json:"ttl" koanf:"ttl"`

	// NegativeTTL is how long a miss is remembered. Keys are short enough to guess at, so
	// without this a scan for keys that do not exist reaches the database on every request.
	// Keep it small: until it lapses, an instance stays blind to a key another instance just
	// created.
	NegativeTTL time.Duration `json:"negative_ttl" koanf:"negative_ttl"`
}

// VisitsConfig tunes visit counting.
//
// Counting a visit synchronously turns every redirect into a write against a single row, which
// makes that row a lock hotspot for exactly the links that are most popular. Buffering trades
// an exact, instantly visible counter — which nobody needs — for one write per batch.
type VisitsConfig struct {
	// Enabled turns buffering off, restoring one write per redirect.
	Enabled bool `json:"enabled" koanf:"enabled"`

	// FlushInterval is how long a counter may sit in memory before being written out. It also
	// bounds what a crash can lose.
	FlushInterval time.Duration `json:"flush_interval" koanf:"flush_interval"`

	// MaxBuffered is how many distinct keys may accumulate before a flush is triggered early,
	// so that a burst does not grow the buffer without bound.
	MaxBuffered int `json:"max_buffered" koanf:"max_buffered"`
}
