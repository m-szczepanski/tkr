package config

import "errors"

// Config is the application configuration model.
// TODO(M1-1): implement full config schema and loader behavior.
type Config struct {
	PollingInterval  string
	MarketHoursOnly  bool
	StaleThreshold   string
	HistoryRetention string
	LogLevel         string
	LogFile          string
	Database         DatabaseConfig
	Providers        ProvidersConfig
	Notifications    NotificationsConfig
}

type DatabaseConfig struct {
	Path string
}

type ProvidersConfig struct {
	Finnhub      ProviderConfig
	Stooq        ProviderConfig
	YahooFinance ProviderConfig
	EODHD        ProviderConfig
}

type ProviderConfig struct {
	APIKey   string
	Enabled  bool
	Priority int
}

type NotificationsConfig struct {
	Email   EmailConfig
	Webhook WebhookConfig
}

type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	From     string
	Password string
	To       []string
}

type WebhookConfig struct {
	URL string
}

// Load loads configuration from disk and environment.
func Load() (Config, error) {
	// TODO(M1-1): implement viper-based configuration loading.
	return Config{}, errors.New("not implemented")
}
