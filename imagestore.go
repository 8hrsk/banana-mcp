package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

// imageCounter gives each saved file a unique, monotonic suffix without needing
// a clock or RNG (both keep the filenames deterministic within a run).
var imageCounter uint64

// mimeToExt maps common image MIME types to file extensions.
func mimeToExt(mime string) string {
	switch strings.ToLower(mime) {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}

// SaveImage decodes a base64 image and writes it to the MCP's temp directory.
// It returns the absolute path to the written file.
func SaveImage(dir, model, b64, mime string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output dir: %w", err)
	}

	n := atomic.AddUint64(&imageCounter, 1)
	safeModel := strings.NewReplacer("/", "-", ":", "-", " ", "-").Replace(model)
	name := fmt.Sprintf("%s-%04d%s", safeModel, n, mimeToExt(mime))
	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, raw, 0644); err != nil {
		return "", fmt.Errorf("failed to write image: %w", err)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return abs, nil
}

// DefaultOutputDir returns the directory where generated images are stored.
// Honors BANANA_MCP_OUTPUT_DIR, otherwise uses <os-temp>/banana-mcp.
func DefaultOutputDir() string {
	if d := os.Getenv("BANANA_MCP_OUTPUT_DIR"); d != "" {
		return d
	}
	return filepath.Join(os.TempDir(), "banana-mcp")
}
