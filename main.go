package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// serverInstructions is surfaced to the agent by the MCP handshake so it knows
// how to use the tools without spending tokens probing them. Keep it short.
const serverInstructions = `banana-mcp generates images with Google Gemini ("nano banana") models.

Workflow:
1. Call get_configuration ONCE to get providers, models, per-model options
   (aspect_ratios, resolutions, thinking_levels) and configured key counts.
   Do NOT re-fetch models yourself — the server already queries Google and caches.
2. If configured_keys is 0, ask the user for a key and call add_api_key.
3. Call generate_image with prompt (+ optional model/aspect_ratio/resolution/
   thinking_level). The image is written to disk; the tool returns its file PATH.
   Pass that path to the user or downstream tools instead of copying image bytes.

Defaults: provider=any configured, model=gemini-3.1-flash-image, aspect_ratio 1:1.
thinking_level applies to *-pro-image models only.`

func main() {
	ks, err := NewKeyStore("keys.json")
	if err != nil {
		log.Fatalf("failed to initialize keystore: %v", err)
	}

	providers := map[string]Provider{
		"ai-studio": NewAIStudioProvider(),
		"vertex":    NewVertexProvider(),
	}

	// Cache model-discovery results for 10 minutes so the agent-facing tools stay
	// fast and don't repeatedly hit Google's API on every call.
	cache := newModelCache(10 * time.Minute)
	outputDir := DefaultOutputDir()

	s := server.NewMCPServer(
		"banana-mcp",
		"1.1.0",
		server.WithToolCapabilities(true),
		server.WithInstructions(serverInstructions),
	)

	// Tool: get_configuration
	getConfTool := mcp.NewTool("get_configuration",
		mcp.WithDescription("Get providers, their models, each model's supported generation options (aspect ratios, resolutions, thinking levels), and configured key counts. Call this ONCE; the server fetches and caches from Google so you don't have to."),
	)
	s.AddTool(getConfTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		type ModelInfo struct {
			ID           string            `json:"id"`
			Capabilities ModelCapabilities `json:"capabilities"`
		}
		type ProviderInfo struct {
			ID             string      `json:"id"`
			Models         []ModelInfo `json:"models"`
			ConfiguredKeys int         `json:"configured_keys"`
		}

		summary := ks.GetConfigSummary()
		var infos []ProviderInfo
		for id, p := range providers {
			pid := id
			prov := p
			models := cache.Models(ctx, pid, func() []string {
				return prov.GetModels(ctx, ks.GetKeys(pid))
			})
			var modelInfos []ModelInfo
			for _, m := range models {
				modelInfos = append(modelInfos, ModelInfo{ID: m, Capabilities: CapabilitiesForModel(m)})
			}
			infos = append(infos, ProviderInfo{
				ID:             pid,
				Models:         modelInfos,
				ConfiguredKeys: summary[pid],
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
		mcp.WithDescription("Generate an image using Gemini image models. Saves the result to a temp file and returns its path (plus the image inline). Use get_configuration to discover valid option values."),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("The image prompt")),
		mcp.WithString("provider", mcp.Description("Optional provider ID. If omitted, tries any provider with configured keys")),
		mcp.WithString("model", mcp.Description("Optional model ID. Defaults to gemini-3.1-flash-image")),
		mcp.WithString("aspect_ratio", mcp.Description("Optional aspect ratio, e.g. 1:1, 16:9, 9:16 (see get_configuration)")),
		mcp.WithString("resolution", mcp.Description("Optional resolution, e.g. 1K, 2K, 4K (model-dependent)")),
		mcp.WithString("thinking_level", mcp.Description("Optional thinking level LOW or HIGH; *-pro-image models only")),
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

		opts := ImageOptions{}
		if val, ok := args["aspect_ratio"].(string); ok {
			opts.AspectRatio = val
		}
		if val, ok := args["resolution"].(string); ok {
			opts.Resolution = strings.ToUpper(val)
		}
		if val, ok := args["thinking_level"].(string); ok {
			opts.ThinkingLevel = strings.ToUpper(val)
		}

		if err := validateOptions(model, opts); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
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
				imgBase64, mimeType, err := p.GenerateImage(ctx, key, model, prompt, opts)
				if err != nil {
					fallbackLogs = append(fallbackLogs, fmt.Sprintf("[%s] Key %d failed: %v", p.ID(), i+1, err))
					continue
				}

				// Persist to disk and hand back the path (token-cheap for the agent).
				path, saveErr := SaveImage(outputDir, model, imgBase64, mimeType)
				if saveErr != nil {
					fallbackLogs = append(fallbackLogs, fmt.Sprintf("[%s] Key %d: generated but save failed: %v", p.ID(), i+1, saveErr))
					continue
				}

				summary := map[string]interface{}{
					"path":     path,
					"provider": p.ID(),
					"model":    model,
					"mime":     mimeType,
					"options":  opts,
				}
				if len(fallbackLogs) > 0 {
					summary["fallback_logs"] = fallbackLogs
				}
				data, _ := json.MarshalIndent(summary, "", "  ")

				result := mcp.NewToolResultText(fmt.Sprintf("Image saved to:\n%s\n\n%s", path, string(data)))
				result.Content = append(result.Content, mcp.ImageContent{
					Type:     "image",
					Data:     imgBase64,
					MIMEType: mimeType,
				})
				return result, nil
			}
		}

		return mcp.NewToolResultError(fmt.Sprintf("All keys failed to generate image.\nLogs:\n%s", strings.Join(fallbackLogs, "\n"))), nil
	})

	log.Printf("Starting banana-mcp server on stdio (output dir: %s)...", outputDir)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
