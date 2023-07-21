package db

import "time"

type Config struct {
	DSN             string        `json:"dsn,omitempty"                koanf:"dsn"`
	Debug           bool          `json:"debug,omitempty"              koanf:"debug"`
	MaxIdelConns    int           `json:"max_idel_conns,omitempty"     koanf:"max_idel_conns"`
	MaxOpenConns    int           `json:"max_open_conns,omitempty"     koanf:"max_open_conns"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time,omitempty" koanf:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime,omitempty"  koanf:"conn_max_lifetime"`
}
