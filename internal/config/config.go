package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	// HTTP
	Port                int `envconfig:"PORT" default:"8080"`
	ReadTimeoutSeconds  int `envconfig:"HTTP_READ_TIMEOUT_SEC" default:"5"`
	WriteTimeoutSeconds int `envconfig:"HTTP_WRITE_TIMEOUT_SEC" default:"5"`
	IdleTimeoutSeconds  int `envconfig:"HTTP_IDLE_TIMEOUT_SEC" default:"30"`

	// Per-request overall timeout for cold path (DB+Redis), in milliseconds
	RequestTimeoutMs int `envconfig:"REQUEST_TIMEOUT_MS" default:"300"`

	// PostgreSQL
	DatabaseURL            string `envconfig:"DATABASE_URL" required:"true"`
	PGMaxConns             int    `envconfig:"PG_MAX_CONNS" default:"24"`
	PGHealthCheckPeriodSec int    `envconfig:"PG_HEALTHCHECK_PERIOD_SEC" default:"30"`
	PGMaxConnIdleTimeSec   int    `envconfig:"PG_MAX_CONN_IDLE_TIME_SEC" default:"300"`
	PGMaxConnLifetimeSec   int    `envconfig:"PG_MAX_CONN_LIFETIME_SEC" default:"0"`

	// Redis
	RedisAddr           string `envconfig:"REDIS_ADDR" required:"true"`
	RedisUsername       string `envconfig:"REDIS_USERNAME" default:""`
	RedisPassword       string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB             int    `envconfig:"REDIS_DB" default:"0"`
	RedisDialTimeout    int    `envconfig:"REDIS_DIAL_TIMEOUT_MS" default:"100"`
	RedisReadTimeout    int    `envconfig:"REDIS_READ_TIMEOUT_MS" default:"100"`
	RedisWriteTimeout   int    `envconfig:"REDIS_WRITE_TIMEOUT_MS" default:"100"`
	RedisPoolSize       int    `envconfig:"REDIS_POOL_SIZE" default:"48"`
	RedisMinIdleConns   int    `envconfig:"REDIS_MIN_IDLE_CONNS" default:"8"`
	RedisPoolTimeoutMs  int    `envconfig:"REDIS_POOL_TIMEOUT_MS" default:"200"`
	RedisMaxRetries     int    `envconfig:"REDIS_MAX_RETRIES" default:"1"`
	RedisConnMaxIdleSec int    `envconfig:"REDIS_CONN_MAX_IDLE_SEC" default:"300"`

	// Cache
	CacheTTLSeconds int `envconfig:"CACHE_TTL_SECONDS" default:"3600"`
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Helpers to expose derived durations
func (c Config) ReadTimeout() time.Duration { return time.Duration(c.ReadTimeoutSeconds) * time.Second }
func (c Config) WriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeoutSeconds) * time.Second
}
func (c Config) IdleTimeout() time.Duration { return time.Duration(c.IdleTimeoutSeconds) * time.Second }
func (c Config) RequestTimeout() time.Duration {
	return time.Duration(c.RequestTimeoutMs) * time.Millisecond
}
func (c Config) CacheTTL() time.Duration { return time.Duration(c.CacheTTLSeconds) * time.Second }
func (c Config) PGHealthCheckPeriod() time.Duration {
	return time.Duration(c.PGHealthCheckPeriodSec) * time.Second
}
func (c Config) PGMaxConnIdleTime() time.Duration {
	return time.Duration(c.PGMaxConnIdleTimeSec) * time.Second
}
func (c Config) PGMaxConnLifetime() time.Duration {
	return time.Duration(c.PGMaxConnLifetimeSec) * time.Second
}
