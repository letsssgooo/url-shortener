package generator

import "context"

const (
	CodeLength = 10
	Alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

type Generator interface {
	Generate(ctx context.Context) (string, error)
}
