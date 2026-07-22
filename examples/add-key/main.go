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
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run add_key.go <provider> <key>")
	}
	provider := os.Args[1]
	key := os.Args[2]

	ctx := context.Background()

	// Start the MCP server process
	c, err := client.NewStdioMCPClient("../../banana-mcp", []string{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer c.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "cli-client", Version: "1.0.0"}

	_, err = c.Initialize(ctx, initReq)
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}

	res, err := c.CallTool(ctx, mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
		Params: mcp.CallToolParams{
			Name: "add_api_key",
			Arguments: map[string]interface{}{
				"provider": provider,
				"key":      key,
			},
		},
	})
	if err != nil {
		log.Fatalf("call failed: %v", err)
	}
	fmt.Printf("Response: %s\n", res.Content[0].(mcp.TextContent).Text)
}
