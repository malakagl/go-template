package config

import (
	"fmt"
	"os"
	"time"

	validate "github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Logging    LoggingConfig    `yaml:"logging"`
	CouponCode CouponCodeConfig `yaml:"couponCode"`
	Telemetry  TelemetryConfig  `yaml:"telemetry"`
}

type TelemetryConfig struct {
	Enabled     bool   `yaml:"enabled"`
	ServiceName string `yaml:"serviceName"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
}

type ServerConfig struct {
	Host                   string        `yaml:"host"`
	Port                   int           `yaml:"port" validate:"required"`
	MaxCouponCodeCacheSize int           `yaml:"maxCouponCodeCacheSize"`
	ReqLimitPerIP          int           `yaml:"reqLimitPerIP" validate:"min=1"`
	ReqBurstPerIP          int           `yaml:"reqBurstPerIP" validate:"min=1"`
	ReqRateWindow          time.Duration `yaml:"reqRateWindow" validate:"min=1m"`
	GracefulTimeout        time.Duration `yaml:"gracefulTimeout" validate:"required"`
}

type DatabaseConfig struct {
	Host                 string        `yaml:"host" validate:"required"`
	Port                 int           `yaml:"port" validate:"required"`
	Name                 string        `yaml:"name" validate:"required"`
	User                 string        `yaml:"user" validate:"required"`
	Password             string        `yaml:"password" validate:"required"`
	MigrationsFolderPath string        `yaml:"migrationsFolderPath" validate:"required"`
	SSLMode              string        `yaml:"sslMode"`
	Debug                bool          `yaml:"debug"`
	Type                 string        `yaml:"type" validate:"required"` // e.g., "postgres", "sqlite"
	MaxOpenConnections   int           `yaml:"maxOpenConnections" validate:"min=1"`
	MaxIdleConnections   int           `yaml:"maxIdleConnections" validate:"min=1"`
	ConnMaxIdleTime      time.Duration `yaml:"connMaxIdleTime" validate:"min=1m"`
	MaxConnMaxLifeTime   time.Duration `yaml:"maxConnMaxLifeTime" validate:"min=1m"`
}

type CouponCodeConfig struct {
	Unzipped  bool     `yaml:"unzipped"`
	FilePaths []string `yaml:"filePaths"`
}

type LoggingConfig struct {
	Level      string `yaml:"level" validate:"required"`
	JsonFormat bool   `yaml:"jsonFormat"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	if err := validate.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err)
	}

	// add defaults
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}

	return cfg, nil
}
