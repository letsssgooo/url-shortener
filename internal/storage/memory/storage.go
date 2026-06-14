package memory

import (
	"context"
	"sync"

	"ozon-intern/internal/storage"
)

type Storage struct {
	mu         sync.RWMutex
	byOriginal map[string]string
	byShort    map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		byOriginal: make(map[string]string),
		byShort:    make(map[string]string),
	}
}

func (s *Storage) Save(ctx context.Context, originalURL string, shortCode string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byOriginal[originalURL]; ok {
		return storage.ErrOriginalURLExists
	}
	if _, ok := s.byShort[shortCode]; ok {
		return storage.ErrShortCodeExists
	}

	s.byOriginal[originalURL] = shortCode
	s.byShort[shortCode] = originalURL

	return nil
}

func (s *Storage) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, ok := s.byShort[shortCode]
	if !ok {
		return "", storage.ErrNotFound
	}

	return originalURL, nil
}

func (s *Storage) GetShortCode(ctx context.Context, originalURL string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	shortCode, ok := s.byOriginal[originalURL]
	if !ok {
		return "", storage.ErrNotFound
	}

	return shortCode, nil
}
