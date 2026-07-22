package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	ks, err := NewKeyStore("keys.json")
	if err != nil {
		log.Fatalf("failed to initialize keystore: %v", err)
	}

	providers := map[string]Provider{
		"ai-studio": NewAIStudioProvider(),
		"vertex":    NewVertexProvider(),
	}

	s := server.NewMCPServer(
		"banana-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Tool: get_configuration
	getConfTool := mcp.NewTool("get_configuration",
		mcp.WithDescription("Get available providers, their supported models, and the count of configured API keys"),
	)
	s.AddTool(getConfTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		type ProviderInfo struct {
			ID              string   `json:"id"`
			SupportedModels []string `json:"supported_models"`
			ConfiguredKeys  int      `json:"configured_keys"`
		}

		summary := ks.GetConfigSummary()
		var infos []ProviderInfo
		for id, p := range providers {
			infos = append(infos, ProviderInfo{
				ID:              id,
				SupportedModels: p.GetModels(ctx, ks.GetKeys(id)),
				ConfiguredKeys:  summary[id],
			})
		}

		data, _ := json.MarshalIndent(infos, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})

	// Tool: add_api_key
	addKeyTool := mcp.NewTool("add_api_key",
		mcp.WithDescription("Add a new API key for a provider"),
		mcp.WithString("provider", mcp.Required(), mcp.Description("Provider ID (e.g. ai-studio, vertex)")),
		mcp.WithString("key", mcp.Required(), mcp.Description("The API key or auth token. For vertex, use 'projectID:region:accessToken'")),
	)
	s.AddTool(addKeyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments.(map[string]interface{})
		provider := args["provider"].(string)
		key := args["key"].(string)

		if _, ok := providers[provider]; !ok {
			return mcp.NewToolResultError(fmt.Sprintf("unknown provider: %s", provider)), nil
		}

		if err := ks.AddKey(provider, key); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save key: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully added key for provider %s", provider)), nil
	})

	// Tool: generate_image
	genImgTool := mcp.NewTool("generate_image",
		mcp.WithDescription("Generate an image using nano banana models (Gemini)"),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("The image prompt")),
		mcp.WithString("provider", mcp.Description("Optional provider ID. If omitted, tries any provider with configured keys")),
		mcp.WithString("model", mcp.Description("Optional model ID. Defaults to gemini-3.1-flash-image")),
	)
	s.AddTool(genImgTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments.(map[string]interface{})
		prompt := args["prompt"].(string)
		
		reqProvider := ""
		if val, ok := args["provider"].(string); ok && val != "" {
			reqProvider = val
		}
		
		model := "gemini-3.1-flash-image"
		if val, ok := args["model"].(string); ok && val != "" {
			model = val
		}

		// Figure out which providers to try
		var providersToTry []Provider
		if reqProvider != "" {
			if p, ok := providers[reqProvider]; ok {
				providersToTry = append(providersToTry, p)
			} else {
				return mcp.NewToolResultError(fmt.Sprintf("unknown provider: %s", reqProvider)), nil
			}
		} else {
			// Add all providers that have at least one key
			for _, p := range providers {
				if len(ks.GetKeys(p.ID())) > 0 {
					providersToTry = append(providersToTry, p)
				}
			}
		}

		if len(providersToTry) == 0 {
			return mcp.NewToolResultError("no providers have configured keys. Use add_api_key first."), nil
		}

		var fallbackLogs []string

		for _, p := range providersToTry {
			keys := ks.GetKeys(p.ID())
			if len(keys) == 0 {
				continue
			}

			for i, key := range keys {
				imgBase64, mimeType, err := p.GenerateImage(ctx, key, model, prompt)
				if err != nil {
					fallbackLogs = append(fallbackLogs, fmt.Sprintf("[%s] Key %d failed: %v", p.ID(), i+1, err))
					continue
				}

				// Success!
				result := mcp.NewToolResultText(fmt.Sprintf("Image generated successfully using provider %s.\nLogs:\n%s", p.ID(), strings.Join(fallbackLogs, "\n")))
				// Add the image content
				result.Content = append(result.Content, mcp.ImageContent{
					Type:     "image",
					Data:     imgBase64,
					MIMEType: mimeType,
				})
				return result, nil
			}
		}

		// If we reach here, all keys failed
		return mcp.NewToolResultError(fmt.Sprintf("All keys failed to generate image.\nLogs:\n%s", strings.Join(fallbackLogs, "\n"))), nil
	})

	log.Println("Starting banana-mcp server on stdio...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
