# banana-mcp — Agent Guide

Image generation via Google Gemini ("nano banana") models. Token-optimized: the
server does provider/model/option discovery and caching for you.

## Tools

| Tool | When to call | Returns |
|------|--------------|---------|
| `get_configuration` | Once at start | Providers, models, per-model options, key counts |
| `add_api_key` | Only if `configured_keys` is 0 | Confirmation |
| `generate_image` | To create an image | File **path** on disk (+ inline image) |

## Flow

1. **`get_configuration`** → read `models[].capabilities` for valid
   `aspect_ratios`, `resolutions`, `thinking_levels`. Do **not** fetch models
   yourself; the server already queried Google and cached the result.
2. If `configured_keys == 0`, ask the user for a key, then **`add_api_key`**
   (`provider`, `key`). For `vertex`, key format is `projectID:region:accessToken`.
3. **`generate_image`**:
   - Required: `prompt`
   - Optional: `provider`, `model` (default `gemini-3.1-flash-image`),
     `aspect_ratio`, `resolution`, `thinking_level`
   - The image is written to a temp file. Use the returned `path`; don't copy the
     raw image bytes around.

## Notes

- `thinking_level` (`LOW`/`HIGH`) works only on `*-pro-image` models.
- Invalid option values are rejected with the allowed set — read `capabilities`
  first to avoid a round-trip.
- Output dir defaults to `<temp>/banana-mcp`; override with
  `BANANA_MCP_OUTPUT_DIR`.
- AI Studio (`ai-studio`) is the primary, fully-tested provider.
