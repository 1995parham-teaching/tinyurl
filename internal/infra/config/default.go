package config

import (
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/1995parham-teaching/tinyurl/internal/infra/http/server"
	"github.com/1995parham-teaching/tinyurl/internal/infra/logger"
	"github.com/1995parham-teaching/tinyurl/internal/infra/repository"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"go.uber.org/fx"
)

// Default return default configuration.
// nolint: mnd, gosec
func Default() Config {
	return Config{
		Out: fx.Out{},
		Generator: generator.Config{
			Type: generator.TypeSimple,
			Key:  "",
		},
		Logger: logger.Config{
			Level: "debug",
		},
		Database: db.Config{
			DSN:             "postgresql://tinyurl:secret@localhost/tinyurl",
			Debug:           true,
			MaxIdelConns:    10,
			MaxOpenConns:    10,
			ConnMaxIdleTime: 10 * time.Second,
			ConnMaxLifetime: 10 * time.Second,
		},
		Repository: repository.Config{
			Cache: repository.CacheConfig{
				Enabled:     true,
				Size:        100_000,
				TTL:         10 * time.Minute,
				NegativeTTL: 5 * time.Second,
			},
			Visits: repository.VisitsConfig{
				Enabled:       true,
				FlushInterval: 5 * time.Second,
				MaxBuffered:   10_000,
			},
		},
		Telemetry: telemetry.Config{
			Namespace:   "1995parham-teaching",
			ServiceName: "tinyurl",
			Meter: telemetry.Meter{
				Address: ":8080",
				Enabled: true,
			},
			Trace: telemetry.Trace{
				Enabled:  false,
				Endpoint: "127.0.0.1:4317",
			},
		},
		Server: server.Config{
			Address: ":1378",
		},
	}
}
