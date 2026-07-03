package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

type ordersArgs struct {
	accountHeaderArgs
}

type dealsArgs struct {
	accountHeaderArgs
}

type historyOrdersArgs struct {
	accountHeaderArgs
	BeginTime string   `json:"begin_time"`
	EndTime   string   `json:"end_time"`
	Codes     []string `json:"codes,omitempty"`
}

type historyDealsArgs struct {
	accountHeaderArgs
	BeginTime string   `json:"begin_time"`
	EndTime   string   `json:"end_time"`
	Codes     []string `json:"codes,omitempty"`
}

// RegisterOrders registers the 4 order history read-only tools on the MCP server.
func RegisterOrders(s *mcp.Server, c moomoo.MoomooClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_orders",
		Description: "List today's open orders for one trading account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ordersArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		orders, err := c.GetOrders(ctx, accountID, args.TrdEnv, args.TrdMarket)
		if err != nil {
			return toolError(fmt.Sprintf("get orders: %v", err)), nil, nil
		}
		return jsonResult(orders), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_deals",
		Description: "List today's filled deals (order fills) for one trading account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args dealsArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		deals, err := c.GetDeals(ctx, accountID, args.TrdEnv, args.TrdMarket)
		if err != nil {
			return toolError(fmt.Sprintf("get deals: %v", err)), nil, nil
		}
		return jsonResult(deals), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_history_orders",
		Description: "List historical orders for one trading account within [begin_time, end_time] (format yyyy-MM-dd HH:mm:ss). codes optionally restricts the result to specific security codes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args historyOrdersArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		orders, err := c.GetHistoryOrders(ctx, accountID, args.TrdEnv, args.TrdMarket, args.BeginTime, args.EndTime, args.Codes)
		if err != nil {
			return toolError(fmt.Sprintf("get history orders: %v", err)), nil, nil
		}
		return jsonResult(orders), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_history_deals",
		Description: "List historical filled deals for one trading account within [begin_time, end_time] (format yyyy-MM-dd HH:mm:ss). codes optionally restricts the result to specific security codes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args historyDealsArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		deals, err := c.GetHistoryDeals(ctx, accountID, args.TrdEnv, args.TrdMarket, args.BeginTime, args.EndTime, args.Codes)
		if err != nil {
			return toolError(fmt.Sprintf("get history deals: %v", err)), nil, nil
		}
		return jsonResult(deals), nil, nil
	})
}
