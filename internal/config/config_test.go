package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("BASE_URL", "")
	t.Setenv("STORAGE_TYPE", "")
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr: %s, want: %s", cfg.HTTPAddr, ":8080")
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Fatalf("BaseURL: %s, want: %s", cfg.BaseURL, "http://localhost:8080")
	}
	if cfg.StorageType != StorageTypeMemory {
		t.Fatalf("StorageType: %s, want: %s", cfg.StorageType, StorageTypeMemory)
	}
}

func TestLoadCustomValues(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("BASE_URL", "https://short.example.com/")
	t.Setenv("STORAGE_TYPE", "postgres")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr: %s, want: %s", cfg.HTTPAddr, ":9090")
	}
	if cfg.BaseURL != "https://short.example.com" {
		t.Fatalf("BaseURL: %s, want: %s", cfg.BaseURL, "https://short.example.com")
	}
	if cfg.StorageType != StorageTypePostgres {
		t.Fatalf("StorageType: %s, want: %s", cfg.StorageType, StorageTypePostgres)
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/db" {
		t.Fatalf("DatabaseURL: %s, want configured database url", cfg.DatabaseURL)
	}
}

func TestLoadInvalidBaseURL(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("BASE_URL", "localhost:8080")
	t.Setenv("STORAGE_TYPE", "")
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadInvalidStorageType(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("BASE_URL", "")
	t.Setenv("STORAGE_TYPE", "unknown")
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadPostgresRequiresDatabaseURL(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("BASE_URL", "")
	t.Setenv("STORAGE_TYPE", "postgres")
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadFromDotEnv(t *testing.T) {
	unsetConfigEnv(t)
	t.Chdir(t.TempDir())

	envContent := []byte("HTTP_ADDR=:9090\nBASE_URL=https://short.example.com\nSTORAGE_TYPE=memory\n")
	if err := os.WriteFile(".env", envContent, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr: %s, want: %s", cfg.HTTPAddr, ":9090")
	}
	if cfg.BaseURL != "https://short.example.com" {
		t.Fatalf("BaseURL: %s, want: %s", cfg.BaseURL, "https://short.example.com")
	}
	if cfg.StorageType != StorageTypeMemory {
		t.Fatalf("StorageType: %s, want: %s", cfg.StorageType, StorageTypeMemory)
	}
}

func TestLoadEnvOverridesDotEnv(t *testing.T) {
	unsetConfigEnv(t)
	t.Chdir(t.TempDir())

	envContent := []byte("HTTP_ADDR=:9090\nBASE_URL=https://from-file.example.com\nSTORAGE_TYPE=memory\n")
	if err := os.WriteFile(".env", envContent, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("BASE_URL", "https://from-env.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "https://from-env.example.com" {
		t.Fatalf("BaseURL: %s, want: %s", cfg.BaseURL, "https://from-env.example.com")
	}
}

func unsetConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{"HTTP_ADDR", "BASE_URL", "STORAGE_TYPE", "DATABASE_URL"}
	for _, key := range keys {
		oldValue, existed := os.LookupEnv(key)

		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%s) error: %v", key, err)
		}

		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(key, oldValue)
				return
			}

			_ = os.Unsetenv(key)
		})
	}
}
