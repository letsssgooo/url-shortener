package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound          = errors.New("link not found")
	ErrOriginalURLExists = errors.New("original url already exists")
	ErrShortCodeExists   = errors.New("short code already exists")
)

type Storage interface {
	Save(ctx context.Context, originalURL string, shortCode string) error
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
	GetShortCode(ctx context.Context, originalURL string) (string, error)
}
