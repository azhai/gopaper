package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/azhai/gopaper/internal/model"

	"github.com/pelletier/go-toml/v2"
)

type AppConfig struct {
	SERVER_PORT     string            `toml:"SERVER_PORT"`
	CONTENT_DIR     string            `toml:"CONTENT_DIR"`
	UPLOAD_DIR      string            `toml:"UPLOAD_DIR"`
	JWT_SECRET      string            `toml:"JWT_SECRET"`
	JWT_TTL         string            `toml:"JWT_TTL"`
	MAX_UPLOAD_SIZE int64             `toml:"MAX_UPLOAD_SIZE"`
	CACHE_SIZE      int               `toml:"CACHE_SIZE"`
	SITE_URL        string            `toml:"SITE_URL"`
	ALLOWED_ORIGINS []string          `toml:"ALLOWED_ORIGINS"`
	Admin           model.AdminConfig `toml:"admin"`
}

func Load(configPath string) (*AppConfig, error) {
	cfg := &AppConfig{
		SERVER_PORT:     "3000",
		CONTENT_DIR:     "./content",
		UPLOAD_DIR:      "./uploads",
		JWT_TTL:         "24h",
		MAX_UPLOAD_SIZE: 5242880,
		CACHE_SIZE:      67108864,
		SITE_URL:        "http://localhost:3000",
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err == nil {
			if err := toml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse config file: %w", err)
			}
		}
	}

	cfg.applyEnvOverrides()

	if cfg.JWT_SECRET == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func (c *AppConfig) applyEnvOverrides() {
	if v := os.Getenv("SERVER_PORT"); v != "" {
		c.SERVER_PORT = v
	}
	if v := os.Getenv("CONTENT_DIR"); v != "" {
		c.CONTENT_DIR = v
	}
	if v := os.Getenv("UPLOAD_DIR"); v != "" {
		c.UPLOAD_DIR = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.JWT_SECRET = v
	}
	if v := os.Getenv("JWT_TTL"); v != "" {
		c.JWT_TTL = v
	}
	if v := os.Getenv("MAX_UPLOAD_SIZE"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.MAX_UPLOAD_SIZE = n
		}
	}
	if v := os.Getenv("CACHE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.CACHE_SIZE = n
		}
	}
	if v := os.Getenv("SITE_URL"); v != "" {
		c.SITE_URL = v
	}
	if v := os.Getenv("ADMIN_USERNAME"); v != "" {
		c.Admin.USERNAME = v
	}
	if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
		c.Admin.PASSWORD = v
	}
}
