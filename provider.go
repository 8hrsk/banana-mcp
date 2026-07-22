package main

import (
	"context"
)

// Provider represents an extensible AI provider for generating images.
type Provider interface {
	// ID returns the unique identifier for the provider (e.g. "vertex", "ai-studio")
	ID() string

	// GenerateImage generates an image from a text prompt using the given key and model.
	// Returns the base64 encoded image, the MIME type, or an error if it failed.
	GenerateImage(ctx context.Context, key string, model string, prompt string) (string, string, error)

	// GetModels returns a list of model names supported by this provider.
	// It can use the provided keys to dynamically fetch the list from the API.
	GetModels(ctx context.Context, keys []string) []string
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
	ResponseModalities []string `json:"responseModalities"`
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

