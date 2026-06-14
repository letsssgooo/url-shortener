package service

import (
	"context"
	"errors"
	"testing"

	"ozon-intern/internal/storage"
	"ozon-intern/internal/storage/memory"
)

type sequenceGenerator struct {
	codes []string
	err   error
	calls int
}

func (g *sequenceGenerator) Generate(ctx context.Context) (string, error) {
	g.calls++
	if g.err != nil {
		return "", g.err
	}
	if len(g.codes) == 0 {
		return "", errors.New("no codes")
	}

	code := g.codes[0]
	g.codes = g.codes[1:]

	return code, nil
}

type originalURLRaceStorage struct {
	shortCode string
}

func (s *originalURLRaceStorage) Save(ctx context.Context, originalURL string, shortCode string) error {
	return storage.ErrOriginalURLExists
}

func (s *originalURLRaceStorage) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	return "", storage.ErrNotFound
}

func (s *originalURLRaceStorage) GetShortCode(ctx context.Context, originalURL string) (string, error) {
	if s.shortCode == "" {
		s.shortCode = "existing"
		return "", storage.ErrNotFound
	}

	return s.shortCode, nil
}

func TestShortenerShortenCreatesShortCode(t *testing.T) {
	storage := memory.NewStorage()
	generator := &sequenceGenerator{codes: []string{"abcDEF123_"}}
	shortener := NewShortener(storage, generator)

	shortCode, err := shortener.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}

	if shortCode != "abcDEF123_" {
		t.Fatalf("Shorten() shortCode: %s, want: %s", shortCode, "abcDEF123_")
	}
}

func TestShortenerShortenReturnsExistingShortCode(t *testing.T) {
	storage := memory.NewStorage()
	generator := &sequenceGenerator{codes: []string{"abcDEF123_", "unused"}}
	shortener := NewShortener(storage, generator)

	firstCode, err := shortener.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Shorten() first error = %v", err)
	}

	secondCode, err := shortener.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Shorten() second error = %v", err)
	}

	if secondCode != firstCode {
		t.Fatalf("Shorten() secondCode: %s, want: %s", secondCode, firstCode)
	}
	if generator.calls != 1 {
		t.Fatalf("Generate() calls = %d, want 1", generator.calls)
	}
}

func TestShortenerShortenRetriesShortCodeCollision(t *testing.T) {
	storage := memory.NewStorage()
	if err := storage.Save(context.Background(), "https://already-saved.example", "collision"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	generator := &sequenceGenerator{codes: []string{"collision", "newCode"}}
	shortener := NewShortener(storage, generator)

	shortCode, err := shortener.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}

	if shortCode != "newCode" {
		t.Fatalf("Shorten() shortCode: %s, want: %s", shortCode, "newCode")
	}
}

func TestShortenerShortenReturnsTooManyCollisions(t *testing.T) {
	storage := memory.NewStorage()
	if err := storage.Save(context.Background(), "https://already-saved.example", "collision"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	codes := make([]string, MaxGenerateAttempts)
	for i := range codes {
		codes[i] = "collision"
	}

	generator := &sequenceGenerator{codes: codes}
	shortener := NewShortener(storage, generator)

	_, err := shortener.Shorten(context.Background(), "https://example.com")
	if !errors.Is(err, ErrTooManyCollisions) {
		t.Fatalf("Shorten() error = %v, want %v", err, ErrTooManyCollisions)
	}
}

func TestShortenerShortenHandlesOriginalURLRace(t *testing.T) {
	storage := &originalURLRaceStorage{}
	generator := &sequenceGenerator{codes: []string{"newCode"}}
	shortener := NewShortener(storage, generator)

	shortCode, err := shortener.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}

	if shortCode != "existing" {
		t.Fatalf("Shorten() shortCode: %s, want: %s", shortCode, "existing")
	}
}

func TestShortenerResolveReturnsOriginalURL(t *testing.T) {
	storage := memory.NewStorage()
	if err := storage.Save(context.Background(), "https://example.com", "abcDEF123_"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	shortener := NewShortener(storage, &sequenceGenerator{})

	originalURL, err := shortener.Resolve(context.Background(), "abcDEF123_")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if originalURL != "https://example.com" {
		t.Fatalf("Resolve() originalURL: %s, want: %s", originalURL, "https://example.com")
	}
}
