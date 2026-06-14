package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const envFile = ".env"

const (
	StorageTypeMemory   = "memory"
	StorageTypePostgres = "postgres"
)

type Config struct {
	HTTPAddr    string
	BaseURL     string
	StorageType string
	DatabaseURL string
}

func Load() (Config, error) {
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		StorageType: getEnv("STORAGE_TYPE", StorageTypeMemory),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	return cfg, nil
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func (c Config) validate() error {
	if strings.TrimSpace(c.HTTPAddr) == "" {
		return errors.New("HTTP_ADDR is empty")
	}

	if err := validateBaseURL(c.BaseURL); err != nil {
		return err
	}

	switch c.StorageType {
	case StorageTypeMemory:
		return nil
	case StorageTypePostgres:
		if strings.TrimSpace(c.DatabaseURL) == "" {
			return errors.New("DATABASE_URL is required for postgres storage")
		}

		return nil
	default:
		return fmt.Errorf("unsupported STORAGE_TYPE: %s", c.StorageType)
	}
}

func validateBaseURL(rawURL string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid BASE_URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("BASE_URL scheme must be http or https")
	}
	if parsedURL.Host == "" {
		return errors.New("BASE_URL host is empty")
	}

	return nil
}
