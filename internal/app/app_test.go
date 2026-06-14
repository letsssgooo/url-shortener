package app

import (
	"context"
	"testing"

	"ozon-intern/internal/config"
	"ozon-intern/internal/storage"
)

func TestNewStorageMemory(t *testing.T) {
	store, cleanup, err := newStorage(context.Background(), config.Config{
		StorageType: config.StorageTypeMemory,
	})
	defer cleanup()

	if err != nil {
		t.Fatalf("newStorage() error = %v", err)
	}
	if store == nil {
		t.Fatal("newStorage() storage = nil, want storage")
	}

	err = store.Save(context.Background(), "https://example.com", "abcDEF123_")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	originalURL, err := store.GetOriginalURL(context.Background(), "abcDEF123_")
	if err != nil {
		t.Fatalf("GetOriginalURL() error = %v", err)
	}
	if originalURL != "https://example.com" {
		t.Fatalf("GetOriginalURL(): %s, want: %s", originalURL, "https://example.com")
	}
}

func TestNewStoragePostgresReturnsErrorWhenPingFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	store, cleanup, err := newStorage(ctx, config.Config{
		StorageType: config.StorageTypePostgres,
		DatabaseURL: "postgres://user:pass@localhost:5432/db",
	})
	defer cleanup()

	if store != nil {
		t.Fatal("newStorage() storage != nil, want nil")
	}
	if err == nil {
		t.Fatal("newStorage() error = nil, want error")
	}
}

func TestNewStorageUnsupportedType(t *testing.T) {
	store, cleanup, err := newStorage(context.Background(), config.Config{
		StorageType: "unknown",
	})
	defer cleanup()

	if store != nil {
		t.Fatal("newStorage() storage != nil, want nil")
	}
	if err == nil {
		t.Fatal("newStorage() error = nil, want error")
	}
}

func TestNewStorageMemoryImplementsStorage(t *testing.T) {
	store, cleanup, err := newStorage(context.Background(), config.Config{
		StorageType: config.StorageTypeMemory,
	})
	defer cleanup()

	if err != nil {
		t.Fatalf("newStorage() error = %v", err)
	}

	var _ storage.Storage = store
}
