package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	ctx := context.Background()

	binaryPath := "./banana-mcp"
	if len(os.Args) > 1 {
		binaryPath = os.Args[1]
	}

	c, err := client.NewStdioMCPClient(binaryPath, []string{})
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

	fmt.Println("MCP Server Initialized")

	// Test generate_image with default model (gemini-3.1-flash-image)
	fmt.Println("\nGenerating image with default model (Nano Banana 2)...")
	res, err := c.CallTool(ctx, mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
		Params: mcp.CallToolParams{
			Name: "generate_image",
			Arguments: map[string]interface{}{
				"prompt":   "A cute banana wearing sunglasses",
				"provider": "ai-studio",
			},
		},
	})
	if err != nil {
		log.Fatalf("call failed: %v", err)
	}

	if res.IsError {
		fmt.Printf("ERROR: %s\n", res.Content[0].(mcp.TextContent).Text)
		os.Exit(1)
	}

	// Check the text content
	fmt.Printf("Text: %s\n", res.Content[0].(mcp.TextContent).Text)

	// Check the image content
	if len(res.Content) > 1 {
		imgContent := res.Content[1].(mcp.ImageContent)
		fmt.Printf("Image MIME: %s\n", imgContent.MIMEType)
		fmt.Printf("Image Data Length: %d bytes\n", len(imgContent.Data))
		fmt.Println("SUCCESS: Image generated!")
	} else {
		fmt.Println("WARNING: No image content in response")
	}
}
