package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AIStudioProvider struct{}

func NewAIStudioProvider() *AIStudioProvider {
	return &AIStudioProvider{}
}

func (p *AIStudioProvider) ID() string {
	return "ai-studio"
}

func (p *AIStudioProvider) SupportedModels() []string {
	return []string{
		"gemini-3.1-flash-image",
		"gemini-3-pro-image",
		"gemini-2.5-flash-image",
	}
}

type PredictRequest struct {
	Instances []PredictInstance `json:"instances"`
	Parameters PredictParams    `json:"parameters"`
}

type PredictInstance struct {
	Prompt string `json:"prompt"`
}

type PredictParams struct {
	SampleCount int `json:"sampleCount"`
}

type PredictResponse struct {
	Predictions []struct {
		BytesBase64Encoded string `json:"bytesBase64Encoded"`
	} `json:"predictions"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func (p *AIStudioProvider) GenerateImage(ctx context.Context, key string, model string, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:predict?key=%s", model, key)

	reqBody := PredictRequest{
		Instances: []PredictInstance{
			{Prompt: prompt},
		},
		Parameters: PredictParams{
			SampleCount: 1,
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

	var apiResp PredictResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API returned error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Predictions) == 0 || apiResp.Predictions[0].BytesBase64Encoded == "" {
		return "", fmt.Errorf("no image returned in response")
	}

	return apiResp.Predictions[0].BytesBase64Encoded, nil
}
