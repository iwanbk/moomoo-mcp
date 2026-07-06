package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hyperjiang/futu/adapt"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

// klTypeMap maps human-readable K-line period names to the SDK int32 constants.
var klTypeMap = map[string]int32{
	"1min":  adapt.KLType_1Min,
	"5min":  adapt.KLType_5Min,
	"15min": adapt.KLType_15Min,
	"30min": adapt.KLType_30Min,
	"60min": adapt.KLType_60Min,
	"day":   adapt.KLType_Day,
	"week":  adapt.KLType_Week,
	"month": adapt.KLType_Month,
}

type snapshotArgs struct {
	Codes []string `json:"codes"`
}

type quoteArgs struct {
	Codes []string `json:"codes"`
}

type klinesArgs struct {
	Code      string `json:"code"`
	KLType    string `json:"kl_type"`
	BeginTime string `json:"begin_time"`
	EndTime   string `json:"end_time"`
}

type orderBookArgs struct {
	Code string `json:"code"`
}

func jsonResult(v any) *mcp.CallToolResult {
	b, _ := json.Marshal(v)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}
}

func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}

// RegisterMarketData registers the 4 market data read-only tools on the MCP server.
func RegisterMarketData(s *mcp.Server, c moomoo.MoomooClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_market_snapshot",
		Description: "Get the current price snapshot for one or more securities (e.g. US.AAPL, HK.00700).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args snapshotArgs) (*mcp.CallToolResult, any, error) {
		snaps, err := c.GetSnapshot(ctx, args.Codes)
		if err != nil {
			return toolError(fmt.Sprintf("get snapshot: %v", err)), nil, nil
		}
		return jsonResult(snaps), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_stock_quote",
		Description: "Get real-time basic quote for one or more securities.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args quoteArgs) (*mcp.CallToolResult, any, error) {
		quotes, err := c.GetQuote(ctx, args.Codes)
		if err != nil {
			return toolError(fmt.Sprintf("get quote: %v", err)), nil, nil
		}
		return jsonResult(quotes), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_historical_klines",
		Description: "Get historical K-line (candlestick) data. kl_type: 1min, 5min, 15min, 30min, 60min, day, week, month. begin_time/end_time format: yyyy-MM-dd. Returns a columnar {columns, rows} object (field names sent once, not per row).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args klinesArgs) (*mcp.CallToolResult, any, error) {
		klType, ok := klTypeMap[args.KLType]
		if !ok {
			return toolError(fmt.Sprintf("unknown kl_type %q; valid values: 1min, 5min, 15min, 30min, 60min, day, week, month", args.KLType)), nil, nil
		}
		klines, err := c.GetKlines(ctx, args.Code, klType, args.BeginTime, args.EndTime)
		if err != nil {
			return toolError(fmt.Sprintf("get klines: %v", err)), nil, nil
		}
		return jsonResult(klinesToColumnar(klines)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_order_book",
		Description: "Get the current bid/ask order book for a security.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args orderBookArgs) (*mcp.CallToolResult, any, error) {
		ob, err := c.GetOrderBook(ctx, args.Code)
		if err != nil {
			return toolError(fmt.Sprintf("get order book: %v", err)), nil, nil
		}
		return jsonResult(ob), nil, nil
	})
}
