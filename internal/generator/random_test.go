package generator

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type errorReader struct {
	err error
}

func (r errorReader) Read(p []byte) (int, error) {
	return 0, r.err
}

func TestRandomGeneratorGenerateLength(t *testing.T) {
	generator := NewRandomGenerator()

	code, err := generator.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(code) != CodeLength {
		t.Fatalf("Generate() length = %d, want %d", len(code), CodeLength)
	}
}

func TestRandomGeneratorGenerateAlphabet(t *testing.T) {
	generator := NewRandomGenerator()

	for i := 0; i < 1000; i++ {
		code, err := generator.Generate(context.Background())
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		for _, char := range code {
			if !strings.ContainsRune(Alphabet, char) {
				t.Fatalf("Generate() returned code: %s with invalid char: %s", code, string(char))
			}
		}
	}
}

func TestRandomGeneratorGenerateContextCanceled(t *testing.T) {
	generator := NewRandomGenerator()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := generator.Generate(ctx)
	if err == nil {
		t.Fatal("Generate() error = nil, want context cancellation error")
	}
}

func TestRandomGeneratorGenerateReaderError(t *testing.T) {
	readerErr := errors.New("read failed")
	generator, err := NewRandomGeneratorWithReader(errorReader{err: readerErr})
	if err != nil {
		t.Fatalf("NewRandomGeneratorWithReader() error = %v", err)
	}

	_, err = generator.Generate(context.Background())
	if !errors.Is(err, readerErr) {
		t.Fatalf("Generate() error = %v, want %v", err, readerErr)
	}
}
