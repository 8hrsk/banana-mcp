package main

import (
	"context"
)

// Provider represents an extensible AI provider for generating images.
type Provider interface {
	// ID returns the unique identifier for the provider (e.g. "vertex", "ai-studio")
	ID() string

	// GenerateImage calls the provider's API with the given key to generate an image.
	// Returns the base64 encoded image, or an error if it failed.
	GenerateImage(ctx context.Context, key string, model string, prompt string) (string, error)

	// SupportedModels returns a list of model names supported by this provider.
	SupportedModels() []string
}
