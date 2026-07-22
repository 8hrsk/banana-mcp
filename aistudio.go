package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type AIStudioProvider struct{}

func NewAIStudioProvider() *AIStudioProvider {
	return &AIStudioProvider{}
}

func (p *AIStudioProvider) ID() string {
	return "ai-studio"
}

func (p *AIStudioProvider) GetModels(ctx context.Context, keys []string) []string {
	defaultModels := []string{
		"gemini-3.1-flash-lite-image",
		"gemini-3.1-flash-image",
		"gemini-3-pro-image",
	}

	if len(keys) == 0 {
		return defaultModels
	}

	for _, key := range keys {
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", key)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var apiResp struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			continue
		}

		var dynamicModels []string
		for _, m := range apiResp.Models {
			name := strings.TrimPrefix(m.Name, "models/")
			if strings.Contains(name, "image") {
				dynamicModels = append(dynamicModels, name)
			}
		}

		if len(dynamicModels) > 0 {
			return dynamicModels
		}
	}

	return defaultModels
}

func (p *AIStudioProvider) GenerateImage(ctx context.Context, key string, model string, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, key)

	reqBody := GenerateContentRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			ResponseModalities: []string{"IMAGE"},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp GenerateContentResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API returned error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 || apiResp.Candidates[0].Content.Parts[0].InlineData == nil {
		return "", fmt.Errorf("no image returned in response")
	}

	return apiResp.Candidates[0].Content.Parts[0].InlineData.Data, nil
}
