// Package config loads and validates app config from env and an optional .env file using Viper.
package config

import (
	"errors"

	"github.com/spf13/viper"
)

// Config holds application configuration loaded from the environment.
type Config struct {
	// GRPCAddr is the address the gRPC server listens on (e.g. :8080).
	GRPCAddr string `mapstructure:"GRPC_ADDR"`
	// DatabaseURL is the Postgres DSN; empty until DB is wired.
	DatabaseURL string `mapstructure:"DATABASE_URL"`
}

// Load reads .env (if present), then builds and validates Config from the environment via Viper.
// Missing .env is ignored (e.g. in CI). Env vars override .env. Returns an error if required fields are invalid.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // ignore ErrConfigFileNotFound

	v.AutomaticEnv()

	v.SetDefault("GRPC_ADDR", ":8080")
	v.SetDefault("DATABASE_URL", "")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.GRPCAddr == "" {
		return nil, errors.New("config: GRPC_ADDR must be set")
	}

	return &cfg, nil
}
