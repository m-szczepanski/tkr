package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/yourname/tkr/internal/apperrors"
)

const defaultConfigPath = "~/.config/tkr/config.yaml"

// Config is the application configuration model.
type Config struct {
	PollingInterval  time.Duration       `mapstructure:"polling_interval"`
	MarketHoursOnly  bool                `mapstructure:"market_hours_only"`
	StaleThreshold   time.Duration       `mapstructure:"stale_threshold"`
	HistoryRetention time.Duration       `mapstructure:"history_retention"`
	LogLevel         string              `mapstructure:"log_level"`
	LogFile          string              `mapstructure:"log_file"`
	Database         DatabaseConfig      `mapstructure:"database"`
	Providers        ProvidersConfig     `mapstructure:"providers"`
	Notifications    NotificationsConfig `mapstructure:"notifications"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type ProvidersConfig struct {
	Finnhub      ProviderConfig `mapstructure:"finnhub"`
	Stooq        ProviderConfig `mapstructure:"stooq"`
	YahooFinance ProviderConfig `mapstructure:"yahoofinance"`
	EODHD        ProviderConfig `mapstructure:"eodhd"`
}

type ProviderConfig struct {
	APIKey   string `mapstructure:"api_key"`
	Enabled  bool   `mapstructure:"enabled"`
	Priority int    `mapstructure:"priority"`
}

type NotificationsConfig struct {
	Email   EmailConfig   `mapstructure:"email"`
	Webhook WebhookConfig `mapstructure:"webhook"`
}

type EmailConfig struct {
	SMTPHost string   `mapstructure:"smtp_host"`
	SMTPPort int      `mapstructure:"smtp_port"`
	From     string   `mapstructure:"from"`
	Password string   `mapstructure:"password"`
	To       []string `mapstructure:"to"`
}

type WebhookConfig struct {
	URL string `mapstructure:"url"`
}

// Load loads configuration from disk and environment.
func Load() (Config, error) {
	v := viper.New()
	setDefaults(v)

	v.SetEnvPrefix("TKR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := bindEnv(v); err != nil {
		return Config{}, fmt.Errorf("bind environment variables: %w", errors.Join(apperrors.ErrConfig, err))
	}

	configPath := os.Getenv("TKR_CONFIG")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	expandedConfigPath, err := expandPath(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("expand config path %q: %w", configPath, errors.Join(apperrors.ErrConfig, err))
	}

	v.SetConfigFile(expandedConfigPath)
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", expandedConfigPath, errors.Join(apperrors.ErrConfig, err))
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", errors.Join(apperrors.ErrConfig, err))
	}

	cfg.LogFile, err = expandPath(cfg.LogFile)
	if err != nil {
		return Config{}, fmt.Errorf("expand log file path %q: %w", cfg.LogFile, errors.Join(apperrors.ErrConfig, err))
	}

	cfg.Database.Path, err = expandPath(cfg.Database.Path)
	if err != nil {
		return Config{}, fmt.Errorf("expand database path %q: %w", cfg.Database.Path, errors.Join(apperrors.ErrConfig, err))
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("polling_interval", "5m")
	v.SetDefault("market_hours_only", true)
	v.SetDefault("stale_threshold", "20m")
	v.SetDefault("history_retention", "2160h")
	v.SetDefault("log_level", "info")
	v.SetDefault("log_file", "~/.local/share/tkr/app.log")
	v.SetDefault("database.path", "~/.local/share/tkr/data.db")

	v.SetDefault("providers.finnhub.enabled", true)
	v.SetDefault("providers.finnhub.priority", 1)
	v.SetDefault("providers.stooq.enabled", true)
	v.SetDefault("providers.stooq.priority", 2)
	v.SetDefault("providers.yahoofinance.enabled", true)
	v.SetDefault("providers.yahoofinance.priority", 3)
	v.SetDefault("providers.eodhd.enabled", true)
	v.SetDefault("providers.eodhd.priority", 4)
}

func bindEnv(v *viper.Viper) error {
	keys := []string{
		"polling_interval",
		"market_hours_only",
		"stale_threshold",
		"history_retention",
		"log_level",
		"log_file",
		"database.path",
		"providers.finnhub.api_key",
		"providers.finnhub.enabled",
		"providers.finnhub.priority",
		"providers.stooq.api_key",
		"providers.stooq.enabled",
		"providers.stooq.priority",
		"providers.yahoofinance.api_key",
		"providers.yahoofinance.enabled",
		"providers.yahoofinance.priority",
		"providers.eodhd.api_key",
		"providers.eodhd.enabled",
		"providers.eodhd.priority",
		"notifications.email.smtp_host",
		"notifications.email.smtp_port",
		"notifications.email.from",
		"notifications.email.password",
		"notifications.email.to",
		"notifications.webhook.url",
	}

	for _, key := range keys {
		if err := v.BindEnv(key); err != nil {
			return err
		}
	}

	return nil
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return homeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:]), nil
	}

	return path, nil
}
