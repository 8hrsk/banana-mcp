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

type VertexProvider struct{}

func NewVertexProvider() *VertexProvider {
	return &VertexProvider{}
}

func (p *VertexProvider) ID() string {
	return "vertex"
}

func (p *VertexProvider) SupportedModels() []string {
	return []string{
		"gemini-3.1-flash-image",
		"gemini-3-pro-image",
		"gemini-2.5-flash-image",
	}
}

func (p *VertexProvider) GenerateImage(ctx context.Context, key string, model string, prompt string) (string, error) {
	// key format: "projectID:region:token"
	parts := strings.SplitN(key, ":", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("vertex api key must be in format 'projectID:region:accessToken'")
	}
	projectID := parts[0]
	region := parts[1]
	token := parts[2]

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict", region, projectID, region, model)

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
	req.Header.Set("Authorization", "Bearer "+token)

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
