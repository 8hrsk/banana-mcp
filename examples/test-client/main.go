package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	ctx := context.Background()

	// Start the MCP server process
	c, err := client.NewStdioMCPClient("../../banana-mcp", []string{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer c.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "test-client", Version: "1.0.0"}

	_, err = c.Initialize(ctx, initReq)
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}

	fmt.Println("✅ MCP Server Initialized")

	// Call add_api_key
	fmt.Println("\n🔧 Adding Dummy Key 1 (AI Studio)...")
	res1, err := c.CallTool(ctx, mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
		Params: mcp.CallToolParams{
			Name: "add_api_key",
			Arguments: map[string]interface{}{
				"provider": "ai-studio",
				"key":      "invalid-key-1",
			},
		},
	})
	if err != nil {
		log.Fatalf("call failed: %v", err)
	}
	fmt.Printf("Response: %s\n", res1.Content[0].(mcp.TextContent).Text)

	fmt.Println("\n🔧 Adding Dummy Key 2 (AI Studio)...")
	c.CallTool(ctx, mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
		Params: mcp.CallToolParams{
			Name: "add_api_key",
			Arguments: map[string]interface{}{
				"provider": "ai-studio",
				"key":      "invalid-key-2",
			},
		},
	})
	fmt.Println("Response: Successfully added key")

	// Call get_configuration
	fmt.Println("\n📊 Getting Configuration...")
	resConfig, err := c.CallTool(ctx, mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
		Params: mcp.CallToolParams{
			Name: "get_configuration",
		},
	})
	if err != nil {
		log.Fatalf("call failed: %v", err)
	}
	fmt.Printf("Config:\n%s\n", resConfig.Content[0].(mcp.TextContent).Text)

	// Call generate_image (should fallback and fail)
	fmt.Println("\n🎨 Generating Image (Should fallback through both keys and return error)...")
	res2, err := c.CallTool(ctx, mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
		Params: mcp.CallToolParams{
			Name: "generate_image",
			Arguments: map[string]interface{}{
				"prompt":   "A photorealistic banana in space",
				"provider": "ai-studio",
			},
		},
	})
	if err != nil {
		log.Fatalf("call failed: %v", err)
	}
	
	if res2.IsError {
		fmt.Printf("Expected Error Output:\n%s\n", res2.Content[0].(mcp.TextContent).Text)
	} else {
		fmt.Println("Unexpectedly succeeded.")
	}
}
