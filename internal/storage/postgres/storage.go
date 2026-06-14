package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ozon-intern/internal/storage"
)

const uniqueViolationCode = "23505"

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) Save(ctx context.Context, originalURL string, shortCode string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO links (original_url, short_code)
		VALUES ($1, $2)
	`, originalURL, shortCode)
	if err != nil {
		return mapError(err)
	}

	return nil
}

func (s *Storage) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	var originalURL string

	err := s.pool.QueryRow(ctx, `
		SELECT original_url
		FROM links
		WHERE short_code = $1
	`, shortCode).Scan(&originalURL)
	if err != nil {
		return "", mapError(err)
	}

	return originalURL, nil
}

func (s *Storage) GetShortCode(ctx context.Context, originalURL string) (string, error) {
	var shortCode string

	err := s.pool.QueryRow(ctx, `
		SELECT short_code
		FROM links
		WHERE original_url = $1
	`, originalURL).Scan(&shortCode)
	if err != nil {
		return "", mapError(err)
	}

	return shortCode, nil
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return storage.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		switch pgErr.ConstraintName {
		case "links_original_url_key":
			return storage.ErrOriginalURLExists
		case "links_short_code_key":
			return storage.ErrShortCodeExists
		}
	}

	return err
}
