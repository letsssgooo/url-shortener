package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"ozon-intern/internal/storage"
)

func TestMapErrorNoRows(t *testing.T) {
	err := mapError(pgx.ErrNoRows)
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("mapError() = %v, want %v", err, storage.ErrNotFound)
	}
}

func TestMapErrorOriginalURLUniqueViolation(t *testing.T) {
	err := mapError(&pgconn.PgError{
		Code:           uniqueViolationCode,
		ConstraintName: "links_original_url_key",
	})
	if !errors.Is(err, storage.ErrOriginalURLExists) {
		t.Fatalf("mapError() = %v, want %v", err, storage.ErrOriginalURLExists)
	}
}

func TestMapErrorShortCodeUniqueViolation(t *testing.T) {
	err := mapError(&pgconn.PgError{
		Code:           uniqueViolationCode,
		ConstraintName: "links_short_code_key",
	})
	if !errors.Is(err, storage.ErrShortCodeExists) {
		t.Fatalf("mapError() = %v, want %v", err, storage.ErrShortCodeExists)
	}
}

func TestMapErrorUnknownError(t *testing.T) {
	sourceErr := errors.New("source error")

	err := mapError(sourceErr)
	if !errors.Is(err, sourceErr) {
		t.Fatalf("mapError() = %v, want %v", err, sourceErr)
	}
}
