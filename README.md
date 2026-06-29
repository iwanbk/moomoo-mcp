# moomoo-mcp

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for the [Moomoo/Futu](https://www.moomoo.com) trading platform, written in Go.

Provides read-only trading tools (market data, account info, order history) via MCP stdio transport, suitable for use with Claude Desktop or any MCP client.

## Prerequisites

1. Download and run **OpenD** from https://www.moomoo.com/download/OpenAPI
2. Log in to OpenD with your Moomoo/Futu account.
3. OpenD listens on `127.0.0.1:11111` by default.

## Configuration

| Environment Variable       | Default       | Description                                              |
|----------------------------|---------------|----------------------------------------------------------|
| `MOOMOO_OPEND_HOST`        | `127.0.0.1`   | OpenD host                                               |
| `MOOMOO_OPEND_PORT`        | `11111`        | OpenD port                                               |
| `MOOMOO_TRADE_PASSWORD`    | –             | Trade password (plain). If unset → SIMULATE-only mode    |
| `MOOMOO_TRADE_PASSWORD_MD5`| –             | Alternative: MD5 of trade password                       |
| `MOOMOO_SECURITY_FIRM`     | –             | e.g. `FUTUSG`, `FUTUINC`, `FUTUHK` — required for REAL  |

## Usage

```bash
# Build
go build -o moomoo-mcp ./cmd/moomoo-mcp

# Run (SIMULATE-only, no trade password set)
./moomoo-mcp

# Run (REAL mode)
export MOOMOO_TRADE_PASSWORD="your-password"
export MOOMOO_SECURITY_FIRM="FUTUSG"
./moomoo-mcp
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "moomoo": {
      "command": "/path/to/moomoo-mcp"
    }
  }
}
```

## License

Apache-2.0 — see [LICENSE](LICENSE).
