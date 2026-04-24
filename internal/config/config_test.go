package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFromFile(t *testing.T) {
	configPath := writeTempConfig(t, `database:
  path: /tmp/tkr-test.db
providers:
  finnhub:
    api_key: file-key
notifications:
  webhook:
    url: https://example.com/webhook
`)

	t.Setenv("TKR_CONFIG", configPath)
	clearEnvOverrides(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.PollingInterval != 5*time.Minute {
		t.Fatalf("PollingInterval = %v, want %v", cfg.PollingInterval, 5*time.Minute)
	}
	if !cfg.MarketHoursOnly {
		t.Fatalf("MarketHoursOnly = %v, want true", cfg.MarketHoursOnly)
	}
	if cfg.StaleThreshold != 20*time.Minute {
		t.Fatalf("StaleThreshold = %v, want %v", cfg.StaleThreshold, 20*time.Minute)
	}
	if cfg.HistoryRetention != 2160*time.Hour {
		t.Fatalf("HistoryRetention = %v, want %v", cfg.HistoryRetention, 2160*time.Hour)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.Database.Path != "/tmp/tkr-test.db" {
		t.Fatalf("Database.Path = %q, want %q", cfg.Database.Path, "/tmp/tkr-test.db")
	}
	if cfg.Providers.Finnhub.APIKey != "file-key" {
		t.Fatalf("Providers.Finnhub.APIKey = %q, want %q", cfg.Providers.Finnhub.APIKey, "file-key")
	}
	if !cfg.Providers.Finnhub.Enabled {
		t.Fatalf("Providers.Finnhub.Enabled = %v, want true", cfg.Providers.Finnhub.Enabled)
	}
	if cfg.Providers.Finnhub.Priority != 1 {
		t.Fatalf("Providers.Finnhub.Priority = %d, want 1", cfg.Providers.Finnhub.Priority)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	configPath := writeTempConfig(t, `log_level: info
database:
  path: /tmp/base.db
`)

	t.Setenv("TKR_CONFIG", configPath)
	clearEnvOverrides(t)
	t.Setenv("TKR_LOG_LEVEL", "debug")
	t.Setenv("TKR_DATABASE_PATH", "/tmp/override.db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.Database.Path != "/tmp/override.db" {
		t.Fatalf("Database.Path = %q, want %q", cfg.Database.Path, "/tmp/override.db")
	}
}

func TestLoadPathExpansion(t *testing.T) {
	configPath := writeTempConfig(t, `log_file: ~/.local/share/tkr/custom.log
database:
  path: ~/.local/share/tkr/custom.db
`)

	t.Setenv("TKR_CONFIG", configPath)
	clearEnvOverrides(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() returned error: %v", err)
	}

	expectedLogFile := filepath.Join(homeDir, ".local", "share", "tkr", "custom.log")
	expectedDBPath := filepath.Join(homeDir, ".local", "share", "tkr", "custom.db")

	if cfg.LogFile != expectedLogFile {
		t.Fatalf("LogFile = %q, want %q", cfg.LogFile, expectedLogFile)
	}
	if cfg.Database.Path != expectedDBPath {
		t.Fatalf("Database.Path = %q, want %q", cfg.Database.Path, expectedDBPath)
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}

	return path
}

func clearEnvOverrides(t *testing.T) {
	t.Helper()

	keys := []string{
		"TKR_POLLING_INTERVAL",
		"TKR_MARKET_HOURS_ONLY",
		"TKR_STALE_THRESHOLD",
		"TKR_HISTORY_RETENTION",
		"TKR_LOG_LEVEL",
		"TKR_LOG_FILE",
		"TKR_DATABASE_PATH",
		"TKR_PROVIDERS_FINNHUB_API_KEY",
		"TKR_PROVIDERS_FINNHUB_ENABLED",
		"TKR_PROVIDERS_FINNHUB_PRIORITY",
		"TKR_PROVIDERS_STOOQ_API_KEY",
		"TKR_PROVIDERS_STOOQ_ENABLED",
		"TKR_PROVIDERS_STOOQ_PRIORITY",
		"TKR_PROVIDERS_YAHOOFINANCE_API_KEY",
		"TKR_PROVIDERS_YAHOOFINANCE_ENABLED",
		"TKR_PROVIDERS_YAHOOFINANCE_PRIORITY",
		"TKR_PROVIDERS_EODHD_API_KEY",
		"TKR_PROVIDERS_EODHD_ENABLED",
		"TKR_PROVIDERS_EODHD_PRIORITY",
		"TKR_NOTIFICATIONS_EMAIL_SMTP_HOST",
		"TKR_NOTIFICATIONS_EMAIL_SMTP_PORT",
		"TKR_NOTIFICATIONS_EMAIL_FROM",
		"TKR_NOTIFICATIONS_EMAIL_PASSWORD",
		"TKR_NOTIFICATIONS_EMAIL_TO",
		"TKR_NOTIFICATIONS_WEBHOOK_URL",
	}

	for _, key := range keys {
		t.Setenv(key, "")
	}
}
