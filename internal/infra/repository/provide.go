package repository

import (
	"github.com/1995parham-teaching/tinyurl/internal/domain/repository/urlrepo"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Provide assembles the repository the service talks to.
//
// The database sits at the bottom; visit buffering wraps it so that counting a redirect no
// longer writes; the cache wraps that so that a redirect usually does not read either. Both
// layers are decorators over the same interface, so switching either off in the configuration
// leaves a shorter stack and changes nothing else.
func Provide(
	cfg Config, lc fx.Lifecycle, urldb *URLDB, tele telemetry.Telemetery, logger *zap.Logger,
) urlrepo.Repository {
	var repo urlrepo.Repository = urldb

	if cfg.Visits.Enabled {
		repo = NewBufferedVisits(cfg.Visits, lc, repo, tele, logger)
	}

	if cfg.Cache.Enabled {
		repo = NewCached(cfg.Cache, repo, tele, logger)
	}

	return repo
}
