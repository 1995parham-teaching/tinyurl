package config

import (
	"time"

	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/1995parham-teaching/tinyurl/internal/infra/logger"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"go.uber.org/fx"
)

// Default return default configuration.
// nolint: gomnd
func Default() Config {
	return Config{
		Out: fx.Out{},
		Generator: generator.Config{
			Type: "simple",
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
	}
}
