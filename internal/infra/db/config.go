package db

import "time"

type Config struct {
	DSN             string        `koanf:"dsn" json:"dsn,omitempty"`
	Debug           bool          `koanf:"debug" json:"debug,omitempty"`
	MaxIdelConns    int           `koanf:"max_idel_conns" json:"max_idel_conns,omitempty"`
	MaxOpenConns    int           `koanf:"max_open_conns" json:"max_open_conns,omitempty"`
	ConnMaxIdleTime time.Duration `koanf:"conn_max_idle_time" json:"conn_max_idle_time,omitempty"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime" json:"conn_max_lifetime,omitempty"`
}
