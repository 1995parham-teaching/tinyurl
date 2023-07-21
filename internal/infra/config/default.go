package config

import (
	"github.com/1989michael/tinyurl/internal/infra/db"
	"github.com/1989michael/tinyurl/internal/infra/telemetry"
)

// Default return default configuration.
func Default() Config {
	return Config{
		Database: db.Config{},
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
