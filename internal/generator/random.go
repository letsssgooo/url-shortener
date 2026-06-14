package generator

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
)

type RandomGenerator struct {
	reader io.Reader
}

func NewRandomGenerator() *RandomGenerator {
	return &RandomGenerator{reader: rand.Reader}
}

func NewRandomGeneratorWithReader(reader io.Reader) (*RandomGenerator, error) {
	if reader == nil {
		return nil, errors.New("reader is nil")
	}

	return &RandomGenerator{reader: reader}, nil
}

func (g *RandomGenerator) Generate(ctx context.Context) (string, error) {
	buffer := make([]byte, CodeLength)

	if err := ctx.Err(); err != nil {
		return "", err
	}

	if _, err := io.ReadFull(g.reader, buffer); err != nil {
		return "", err
	}

	code := make([]byte, CodeLength)
	for i, value := range buffer {
		code[i] = Alphabet[int(value)%len(Alphabet)]
	}

	return string(code), nil
}
