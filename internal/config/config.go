package config

import (
	"os"
)

type Config struct {
	ApiToken  string
	AccountID string
	ListID    string
	LogPath   string
	DBPath    string
}

func NewAppConfig() (Config, error) {
	cfg := Config{
		ApiToken:  os.Getenv("CF_API_TOKEN"),
		AccountID: os.Getenv("CF_ACCOUNT_ID"),
		ListID:    os.Getenv("CF_LIST_ID"),
		LogPath:   os.Getenv("LOG_PATH"),
		DBPath:    os.Getenv("DB_PATH"),
	}
	cfg = withDefaults(cfg)
	return cfg, nil
}

func withDefaults(cfg Config) Config {
	if cfg.DBPath == "" {
		cfg.DBPath = "/var/lib/fail2cloudflare/fail2cloudflare.db"
	}
	if cfg.LogPath == "" {
		cfg.LogPath = "/var/log/fail2cloudflare/fail2cloudflare.log"
	}
	return cfg
}
