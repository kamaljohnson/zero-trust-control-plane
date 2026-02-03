// Package config loads and validates app config from env and an optional .env file using Viper.
package config

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds application configuration loaded from the environment.
type Config struct {
	// GRPCAddr is the address the gRPC server listens on (e.g. :8080).
	GRPCAddr string `mapstructure:"GRPC_ADDR"`
	// DatabaseURL is the Postgres DSN; empty until DB is wired.
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	// JWTPrivateKey is the PEM-encoded private key (RSA or ECDSA) or path to file; used with JWT_PUBLIC_KEY for RS256/ES256.
	JWTPrivateKey string `mapstructure:"JWT_PRIVATE_KEY"`
	// JWTPublicKey is the PEM-encoded public key or path to file; used with JWT_PRIVATE_KEY.
	JWTPublicKey string `mapstructure:"JWT_PUBLIC_KEY"`
	// JWTIssuer is the iss claim (e.g. "ztcp-auth"); required when auth is enabled.
	JWTIssuer string `mapstructure:"JWT_ISSUER"`
	// JWTAudience is the aud claim (e.g. "ztcp-api"); required when auth is enabled.
	JWTAudience string `mapstructure:"JWT_AUDIENCE"`
	// JWTAccessTTL is the access token lifetime (e.g. "15m"). Used when auth is enabled.
	JWTAccessTTL string `mapstructure:"JWT_ACCESS_TTL"`
	// JWTRefreshTTL is the refresh token lifetime (e.g. "7d"). Used when auth is enabled.
	JWTRefreshTTL string `mapstructure:"JWT_REFRESH_TTL"`
	// BcryptCost is the bcrypt cost factor (4–31); default 12. Used when auth is enabled.
	BcryptCost int `mapstructure:"BCRYPT_COST"`
	// SMSLocalAPIKey is the API key for SMS Local (PoC MFA OTP). Required when MFA is required and no fallback.
	SMSLocalAPIKey string `mapstructure:"SMS_LOCAL_API_KEY"`
	// SMSLocalSender is the optional sender ID for SMS Local.
	SMSLocalSender string `mapstructure:"SMS_LOCAL_SENDER"`
	// SMSLocalBaseURL is the SMS Local API base URL (default https://www.smslocal.com/dev/bulkV2).
	SMSLocalBaseURL string `mapstructure:"SMS_LOCAL_BASE_URL"`
	// DefaultTrustTTLDays is the default device trust TTL in days when platform_settings has no value (e.g. 30).
	DefaultTrustTTLDays int `mapstructure:"DEFAULT_TRUST_TTL_DAYS"`
	// OTPReturnToClient when true enables dev OTP mode: no SMS, OTP stored for GET /dev/mfa/otp; for PoC without DLT. Must not be true when Env is production (panic at startup).
	OTPReturnToClient bool `mapstructure:"OTP_RETURN_TO_CLIENT"`
	// Env is the application environment (e.g. "development", "production"). Used with OTPReturnToClient to panic if dev OTP is enabled in production.
	Env string `mapstructure:"APP_ENV"`

	// Telemetry (optional). When Kafka brokers are set, gRPC server emits telemetry to Kafka.
	// TelemetryKafkaBrokers is a comma-separated list of Kafka broker addresses (e.g. "localhost:9092").
	TelemetryKafkaBrokers string `mapstructure:"KAFKA_BROKERS"`
	// TelemetryKafkaTopic is the Kafka topic for telemetry events (default ztcp-telemetry).
	TelemetryKafkaTopic string `mapstructure:"TELEMETRY_KAFKA_TOPIC"`

	// Worker-only: Loki URL for the telemetry worker to push logs (e.g. http://localhost:3100).
	LokiURL string `mapstructure:"LOKI_URL"`
	// KafkaGroupID is the consumer group ID for the telemetry worker.
	KafkaGroupID string `mapstructure:"KAFKA_GROUP_ID"`
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
	v.SetDefault("JWT_ISSUER", "ztcp-auth")
	v.SetDefault("JWT_AUDIENCE", "ztcp-api")
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "168h") // 7d
	v.SetDefault("BCRYPT_COST", 12)
	v.SetDefault("SMS_LOCAL_BASE_URL", "https://app.smslocal.in/api/smsapi")
	v.SetDefault("DEFAULT_TRUST_TTL_DAYS", 30)
	v.SetDefault("OTP_RETURN_TO_CLIENT", false)
	v.SetDefault("APP_ENV", "")
	v.SetDefault("TELEMETRY_KAFKA_TOPIC", "ztcp-telemetry")
	v.SetDefault("KAFKA_BROKERS", "")
	v.SetDefault("LOKI_URL", "")
	v.SetDefault("KAFKA_GROUP_ID", "ztcp-telemetry-worker")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.GRPCAddr == "" {
		return nil, errors.New("config: GRPC_ADDR must be set")
	}

	if cfg.OTPReturnToClient && cfg.Env == "production" {
		return nil, errors.New("config: OTP_RETURN_TO_CLIENT must not be true when APP_ENV=production")
	}

	if cfg.BcryptCost == 0 {
		cfg.BcryptCost = 12
	}
	if cfg.BcryptCost < 4 || cfg.BcryptCost > 31 {
		return nil, errors.New("config: BCRYPT_COST must be between 4 and 31")
	}

	return &cfg, nil
}

// AccessTTL parses JWTAccessTTL as a time.Duration. Returns 15m if unset or invalid.
func (c *Config) AccessTTL() time.Duration {
	d, err := time.ParseDuration(c.JWTAccessTTL)
	if err != nil || d <= 0 {
		return 15 * time.Minute
	}
	return d
}

// RefreshTTL parses JWTRefreshTTL as a time.Duration. Returns 168h if unset or invalid.
func (c *Config) RefreshTTL() time.Duration {
	d, err := time.ParseDuration(c.JWTRefreshTTL)
	if err != nil || d <= 0 {
		return 168 * time.Hour
	}
	return d
}

// TelemetryKafkaBrokersList returns Kafka broker addresses from the comma-separated config.
// Used to decide if telemetry is enabled (non-empty list) and to create the producer.
func (c *Config) TelemetryKafkaBrokersList() []string {
	if c == nil || c.TelemetryKafkaBrokers == "" {
		return nil
	}
	parts := strings.Split(c.TelemetryKafkaBrokers, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
