package main

import (
	"context"
)

// Provider represents an extensible AI provider for generating images.
type Provider interface {
	// ID returns the unique identifier for the provider (e.g. "vertex", "ai-studio")
	ID() string

	// GenerateImage generates an image from a text prompt using the given key, model
	// and generation options. Returns the base64 encoded image, the MIME type, or an
	// error if it failed.
	GenerateImage(ctx context.Context, key string, model string, prompt string, opts ImageOptions) (string, string, error)

	// GetModels returns a list of model names supported by this provider.
	// It can use the provided keys to dynamically fetch the list from the API.
	GetModels(ctx context.Context, keys []string) []string
}

// ImageOptions holds the tunable generation parameters for an image request.
// Empty fields are omitted from the API request (provider defaults apply).
type ImageOptions struct {
	AspectRatio   string // e.g. "1:1", "16:9", "9:16"
	Resolution    string // e.g. "1K", "2K", "4K"
	ThinkingLevel string // e.g. "LOW", "HIGH" (pro image models only)
}

type GenerateContentRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GenerationConfig struct {
	ResponseModalities []string        `json:"responseModalities"`
	ImageConfig        *ImageConfig    `json:"imageConfig,omitempty"`
	ThinkingConfig     *ThinkingConfig `json:"thinkingConfig,omitempty"`
}

type ImageConfig struct {
	AspectRatio string `json:"aspectRatio,omitempty"`
	ImageSize   string `json:"imageSize,omitempty"`
}

type ThinkingConfig struct {
	ThinkingLevel string `json:"thinkingLevel,omitempty"`
}

// buildGenerationConfig assembles the generationConfig payload from options.
func buildGenerationConfig(opts ImageOptions) GenerationConfig {
	cfg := GenerationConfig{ResponseModalities: []string{"IMAGE"}}

	if opts.AspectRatio != "" || opts.Resolution != "" {
		cfg.ImageConfig = &ImageConfig{
			AspectRatio: opts.AspectRatio,
			ImageSize:   opts.Resolution,
		}
	}
	if opts.ThinkingLevel != "" {
		cfg.ThinkingConfig = &ThinkingConfig{ThinkingLevel: opts.ThinkingLevel}
	}
	return cfg
}

type GenerateContentResponse struct {
	Candidates []Candidate `json:"candidates"`
	Error      *APIError   `json:"error,omitempty"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

