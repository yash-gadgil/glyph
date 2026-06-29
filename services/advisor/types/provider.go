package types

import (
	"context"
)

type Provider interface {
	Stream(ctx context.Context, system, prompt string, emit func(string) error) error

	CompleteShort(ctx context.Context, system, prompt string, maxTokens int32) (string, error)
}
