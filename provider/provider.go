package provider

import (
	"context"
	"io"
)

// Provider represents Gemini LLM provider.
type Provider interface {
	Close() error
	CurrentModel() string
	ListModels(ctx context.Context) ([]string, error)
	Generate(ctx context.Context, prompt string, w io.Writer) error
}
