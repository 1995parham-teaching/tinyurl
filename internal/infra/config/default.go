package config

import (
	"time"

	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/1989michael/tinyurl/internal/infra/logger"
	"github.com/1989michael/tinyurl/internal/infra/telemetry"
)

// Default return default configuration.
// nolint: gomnd
func Default() Config {
	return Config{
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
			Namespace:   "1989michael",
			ServiceName: "tinyurl",
			Meter: telemetry.Meter{
				Address: ":8080",
				Enabled: true,
			},
			Trace: telemetry.Trace{
				Enabled: false,
				Agent: telemetry.Agent{
					Port: "6831",
					Host: "127.0.0.1",
				},
			},
		},
	}
}
