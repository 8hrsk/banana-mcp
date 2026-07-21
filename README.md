# Banana MCP Server

Banana MCP Server is a Model Context Protocol (MCP) server written in Go designed to provide AI agents with image generation capabilities utilizing Google's Gemini models (often referred to as "Nano Banana" models in the community).

The server is built with a focus on speed, reliability, and security. It features a pluggable architecture, allowing easy integration with multiple AI image generation providers.

## Key Features

- **Multi-Key Routing and Fallback:** Configure multiple API keys for a single provider. If a generation request fails due to rate limiting, invalid keys, or network errors, the server gracefully falls back to the next available key without interrupting the agent's workflow.
- **Dynamic Key Management:** AI agents can securely add new API keys via the MCP `add_api_key` tool on the fly, eliminating the need to hardcode keys in environment variables. Keys are securely stored in a local `keys.json` file.
- **Pluggable Architecture:** The codebase uses a generic `Provider` interface. Adding support for new providers (like HuggingFace, OpenRouter, etc.) is as simple as implementing the interface and registering it.
- **Supported Providers out of the box:**
  - Google AI Studio (`ai-studio`)
  - Google Cloud Vertex AI (`vertex`)

## Available MCP Tools

Once connected to an MCP client, the following tools become available to the AI agent:

- `get_configuration`: Returns a list of available providers, their supported models, and the number of API keys currently configured for each provider.
- `add_api_key`: Accepts a provider ID and an API key string, saving it securely for future generations.
- `generate_image`: Takes a natural language prompt and optional model/provider arguments to generate an image. Returns the base64-encoded image along with a log of any key fallback events that occurred during the process.

## Installation and Setup

### Prerequisites

- Go 1.21 or higher

### Building from Source

1. Clone the repository:
   ```bash
   git clone git@github.com:8hrsk/banana-mcp.git
   cd banana-mcp
   ```

2. Build the server binary:
   ```bash
   go build -o banana-mcp .
   ```

### Connecting to an MCP Client

Configure your MCP client (such as Claude Desktop or the Antigravity IDE) to execute the compiled binary. For example:

```json
{
  "mcpServers": {
    "banana": {
      "command": "/absolute/path/to/banana-mcp",
      "args": []
    }
  }
}
```

## Adding API Keys

### Google AI Studio
Using an agent or directly invoking the `add_api_key` tool, provide the `ai-studio` provider identifier and your standard Gemini API key.

### Google Cloud Vertex AI
For Vertex AI, you must provide the key in the following format:
`projectID:region:accessToken`
Example: `my-gcp-project:us-central1:ya29.c.c0AY...`

## Extending the Server

To add a new provider:
1. Create a new Go file (e.g., `myprovider.go`).
2. Implement the `Provider` interface defined in `provider.go`.
3. Register your new provider in the `providers` map located in `main.go`.
4. Rebuild the server.

## License

MIT License
