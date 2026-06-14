package service

import (
	"context"
	"errors"

	"ozon-intern/internal/generator"
	"ozon-intern/internal/storage"
)

var (
	ErrTooManyCollisions = errors.New("too many short code collisions")
)

const MaxGenerateAttempts = 10

type Shortener struct {
	storage   storage.Storage
	generator generator.Generator
}

func NewShortener(storage storage.Storage, generator generator.Generator) *Shortener {
	return &Shortener{
		storage:   storage,
		generator: generator,
	}
}

func (s *Shortener) Shorten(ctx context.Context, originalURL string) (string, error) {
	shortCode, err := s.storage.GetShortCode(ctx, originalURL)
	if err == nil {
		return shortCode, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return "", err
	}

	for attempt := 0; attempt < MaxGenerateAttempts; attempt++ {
		shortCode, err = s.generator.Generate(ctx)
		if err != nil {
			return "", err
		}

		err = s.storage.Save(ctx, originalURL, shortCode)
		if err == nil {
			return shortCode, nil
		}

		if errors.Is(err, storage.ErrShortCodeExists) {
			continue
		}

		if errors.Is(err, storage.ErrOriginalURLExists) {
			return s.storage.GetShortCode(ctx, originalURL)
		}

		return "", err
	}

	return "", ErrTooManyCollisions
}

func (s *Shortener) Resolve(ctx context.Context, shortCode string) (string, error) {
	return s.storage.GetOriginalURL(ctx, shortCode)
}
