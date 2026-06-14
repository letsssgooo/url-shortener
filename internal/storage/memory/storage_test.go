package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	storagepkg "ozon-intern/internal/storage"
)

func TestStorageSaveAndGet(t *testing.T) {
	store := NewStorage()

	err := store.Save(context.Background(), "https://example.com", "abcDEF123_")
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

	shortCode, err := store.GetShortCode(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("GetShortCode() error = %v", err)
	}
	if shortCode != "abcDEF123_" {
		t.Fatalf("GetShortCode(): %s, want: %s", shortCode, "abcDEF123_")
	}
}

func TestStorageSaveOriginalURLExists(t *testing.T) {
	store := NewStorage()

	if err := store.Save(context.Background(), "https://example.com", "firstCode"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err := store.Save(context.Background(), "https://example.com", "secondCode")
	if !errors.Is(err, storagepkg.ErrOriginalURLExists) {
		t.Fatalf("Save() error = %v, want %v", err, storagepkg.ErrOriginalURLExists)
	}
}

func TestStorageSaveShortCodeExists(t *testing.T) {
	store := NewStorage()

	if err := store.Save(context.Background(), "https://first.example", "sameCode"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err := store.Save(context.Background(), "https://second.example", "sameCode")
	if !errors.Is(err, storagepkg.ErrShortCodeExists) {
		t.Fatalf("Save() error = %v, want %v", err, storagepkg.ErrShortCodeExists)
	}
}

func TestStorageGetOriginalURLNotFound(t *testing.T) {
	store := NewStorage()

	_, err := store.GetOriginalURL(context.Background(), "missing")
	if !errors.Is(err, storagepkg.ErrNotFound) {
		t.Fatalf("GetOriginalURL() error = %v, want %v", err, storagepkg.ErrNotFound)
	}
}

func TestStorageGetShortCodeNotFound(t *testing.T) {
	store := NewStorage()

	_, err := store.GetShortCode(context.Background(), "https://missing.example")
	if !errors.Is(err, storagepkg.ErrNotFound) {
		t.Fatalf("GetShortCode() error = %v, want %v", err, storagepkg.ErrNotFound)
	}
}

func TestStorageContextCanceled(t *testing.T) {
	store := NewStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := store.Save(ctx, "https://example.com", "abcDEF123_"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Save() error = %v, want %v", err, context.Canceled)
	}

	if _, err := store.GetOriginalURL(ctx, "abcDEF123_"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetOriginalURL() error = %v, want %v", err, context.Canceled)
	}

	if _, err := store.GetShortCode(ctx, "https://example.com"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetShortCode() error = %v, want %v", err, context.Canceled)
	}
}

func TestStorageConcurrentAccess(t *testing.T) {
	store := NewStorage()
	const workers = 100

	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			originalURL := fmt.Sprintf("https://example.com/%d", i)
			shortCode := fmt.Sprintf("code%06d", i)

			if err := store.Save(context.Background(), originalURL, shortCode); err != nil {
				errCh <- fmt.Errorf("save %d: %w", i, err)
				return
			}

			gotOriginalURL, err := store.GetOriginalURL(context.Background(), shortCode)
			if err != nil {
				errCh <- fmt.Errorf("get original %d: %w", i, err)
				return
			}
			if gotOriginalURL != originalURL {
				errCh <- fmt.Errorf("get original %d: %s, want: %s", i, gotOriginalURL, originalURL)
				return
			}

			gotShortCode, err := store.GetShortCode(context.Background(), originalURL)
			if err != nil {
				errCh <- fmt.Errorf("get short %d: %w", i, err)
				return
			}
			if gotShortCode != shortCode {
				errCh <- fmt.Errorf("get short %d: %s, want: %s", i, gotShortCode, shortCode)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}
