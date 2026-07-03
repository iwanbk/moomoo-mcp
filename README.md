# moomoo-mcp

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for the [Moomoo/Futu](https://www.moomoo.com) trading platform, written in Go.

Provides read-only trading tools (system health, market data, account info) via MCP stdio transport, suitable for use with Claude Desktop or any MCP client.

## Available tools

| Done | Domain      | Tool                     | Description                                                        |
|:----:|-------------|--------------------------|----------------------------------------------------------------------|
| ✅   | System      | `check_health`           | Check connectivity to the OpenD gateway.                            |
| ✅   | Market data | `get_stock_quote`        | Real-time basic quote for one or more securities.                   |
| ✅   | Market data | `get_market_snapshot`    | Current price snapshot for one or more securities.                  |
| ✅   | Market data | `get_historical_klines`  | Historical K-line (candlestick) data.                               |
| ✅   | Market data | `get_order_book`         | Current bid/ask order book for a security.                          |
| ✅   | Account     | `get_accounts`           | List trading accounts visible to this OpenD session.                |
| ✅   | Account     | `get_assets`             | Funds/asset info (cash, buying power, market value) for an account. |
| ✅   | Account     | `get_positions`          | Open positions for an account.                                      |
| ✅   | Account     | `get_account_summary`    | Combined assets + positions summary for an account.                 |
| ✅   | Account     | `get_max_tradable`       | Maximum tradable quantities for a security under an account.        |
| ✅   | Account     | `get_margin_ratio`       | Margin ratio info for one or more securities under an account.      |
| ✅   | Account     | `get_cash_flow`          | Trading cash flow summary for an account on a given date.           |
| ✅   | Watchlist   | `get_user_security_group`| List the user's watchlist groups.                                   |
| ✅   | Watchlist   | `get_user_security`      | List securities within a watchlist group.                           |
| ⬜   | Order history | `get_orders`            | Today's orders for an account.                                      |
| ⬜   | Order history | `get_deals`              | Today's filled deals for an account.                                 |
| ⬜   | Order history | `get_history_orders`     | Historical orders for an account.                                    |
| ⬜   | Order history | `get_history_deals`      | Historical filled deals for an account.                              |
| ⬜   | Trading     | `unlock_trade`           | Unlock a REAL account for trading (uses `MOOMOO_TRADE_PASSWORD`/`_MD5` and `MOOMOO_SECURITY_FIRM`). |
| ⬜   | Trading     | `place_order` / `modify_order` / `cancel_order` | Submit, change, or cancel an order.                  |

> This server is read-only today — no order can be placed. Trading tools are planned for a later phase.

## Prerequisites

1. Download and run **OpenD** from https://www.moomoo.com/download/OpenAPI
2. Log in to OpenD with your Moomoo/Futu account.
3. OpenD listens on `127.0.0.1:11111` by default.

## Configuration

| Environment Variable       | Default       | Description                                              |
|----------------------------|---------------|----------------------------------------------------------|
| `MOOMOO_OPEND_HOST`        | `127.0.0.1`   | OpenD host                                               |
| `MOOMOO_OPEND_PORT`        | `11111`        | OpenD port                                               |
| `MOOMOO_TRADE_PASSWORD`    | –             | Safety gate for REAL-account access (see below). Its value is not sent to OpenD yet — trading is not implemented. |
| `MOOMOO_TRADE_PASSWORD_MD5`| –             | MD5 form of `MOOMOO_TRADE_PASSWORD`; same effect.        |
| `MOOMOO_SECURITY_FIRM`     | –             | Not used yet; reserved for the future `unlock_trade` tool. |

By default (no trade password set) the server is **SIMULATE-only**: account tools (`get_assets`, `get_positions`, etc.) can only query SIMULATE accounts, and requests with `trd_env=REAL` are rejected. Setting `MOOMOO_TRADE_PASSWORD` (or `_MD5`) lifts this gate so the read-only account tools can also query REAL accounts — it does not unlock trading itself. Once trading tools are implemented, `MOOMOO_TRADE_PASSWORD`/`_MD5` and `MOOMOO_SECURITY_FIRM` will also be used to authorize `unlock_trade` on a REAL account.

## Usage

```bash
# Build
go build -o moomoo-mcp ./cmd/moomoo-mcp

# Run (SIMULATE-only, no trade password set)
./moomoo-mcp

# Run (allow REAL-account queries)
export MOOMOO_TRADE_PASSWORD="your-password"
export MOOMOO_SECURITY_FIRM="FUTUSG"
./moomoo-mcp
```

### Claude Desktop

Claude Desktop is officially available on macOS and Windows. Config file location:

| OS      | Path                                                       |
|---------|--------------------------------------------------------------|
| macOS   | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json`                |

Add the `moomoo` entry under `mcpServers`:

```json
{
  "mcpServers": {
    "moomoo": {
      "command": "/path/to/moomoo-mcp",
      "env": {
        "MOOMOO_OPEND_HOST": "127.0.0.1",
        "MOOMOO_OPEND_PORT": "11111"
      }
    }
  }
}
```

Restart Claude Desktop after saving.

### Claude Code CLI

Use `claude mcp add` to register the server. The `-s` flag controls where the config is stored:

```bash
# User-level (available in all projects, stored in ~/.claude.json)
claude mcp add -s user moomoo /path/to/moomoo-mcp

# Project-level (stored in .mcp.json in the project directory, can be shared via git)
claude mcp add -s project moomoo /path/to/moomoo-mcp
```

With environment variables (e.g. to allow REAL-account queries):

```bash
claude mcp add -s user moomoo \
  -e MOOMOO_TRADE_PASSWORD=your-password \
  -e MOOMOO_SECURITY_FIRM=FUTUSG \
  -- /path/to/moomoo-mcp
```

Verify the server is registered and healthy:

```bash
claude mcp list
claude mcp get moomoo
```

## License

Apache-2.0 — see [LICENSE](LICENSE).
