package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string

	OpenAIKey string
	BraveKey  string

	APNs APNsConfig
}

type APNsConfig struct {
	KeyPath    string
	KeyID      string
	TeamID     string
	BundleID   string
	Production bool
}

// Load reads from .env (if present) and env vars. Returns error on missing required.
func Load() (Config, error) {
	_ = godotenv.Load()
	c := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		HTTPAddr:    or(os.Getenv("HTTP_ADDR"), ":8080"),
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		BraveKey:    os.Getenv("BRAVE_SEARCH_API_KEY"),
		APNs: APNsConfig{
			KeyPath:    os.Getenv("APNS_KEY_PATH"),
			KeyID:      os.Getenv("APNS_KEY_ID"),
			TeamID:     os.Getenv("APNS_TEAM_ID"),
			BundleID:   os.Getenv("APNS_BUNDLE_ID"),
			Production: os.Getenv("APNS_PRODUCTION") == "true",
		},
	}
	if c.DatabaseURL == "" {
		return c, errors.New("DATABASE_URL required")
	}
	if c.OpenAIKey == "" {
		return c, errors.New("OPENAI_API_KEY required")
	}
	if c.BraveKey == "" {
		return c, errors.New("BRAVE_SEARCH_API_KEY required")
	}
	return c, nil
}

func (a APNsConfig) Configured() bool {
	return a.KeyPath != "" && a.KeyID != "" && a.TeamID != "" && a.BundleID != ""
}

func or(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
