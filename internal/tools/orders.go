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
	Fields    []string `json:"fields,omitempty" jsonschema:"Optional list of output columns, in the order they should appear. Valid values: order_id, order_id_ex, code, name, trd_side, order_type, order_status, qty, price, fill_qty, fill_avg_price, create_time, update_time, remark, last_err_msg. Default when omitted: order_id, code, name, trd_side, order_type, order_status, qty, price, fill_qty, fill_avg_price, create_time, update_time."`
}

type historyDealsArgs struct {
	accountHeaderArgs
	BeginTime string   `json:"begin_time"`
	EndTime   string   `json:"end_time"`
	Codes     []string `json:"codes,omitempty"`
	Fields    []string `json:"fields,omitempty" jsonschema:"Optional list of output columns, in the order they should appear. Valid values: fill_id, fill_id_ex, order_id, order_id_ex, code, name, trd_side, qty, price, create_time, status, counter_broker_name. Default when omitted: fill_id, order_id, code, name, trd_side, qty, price, create_time, status."`
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
		Description: "List historical orders for one trading account within [begin_time, end_time] (format yyyy-MM-dd HH:mm:ss). codes optionally restricts the result to specific security codes. Returns a columnar {columns, rows} object (field names sent once, not per row). Use `fields` to select which columns are returned; defaults to a lean subset.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args historyOrdersArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		orders, err := c.GetHistoryOrders(ctx, accountID, args.TrdEnv, args.TrdMarket, args.BeginTime, args.EndTime, args.Codes)
		if err != nil {
			return toolError(fmt.Sprintf("get history orders: %v", err)), nil, nil
		}
		col, err := ordersToColumnar(orders, args.Fields)
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		return jsonResult(col), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_history_deals",
		Description: "List historical filled deals for one trading account within [begin_time, end_time] (format yyyy-MM-dd HH:mm:ss). codes optionally restricts the result to specific security codes. Returns a columnar {columns, rows} object (field names sent once, not per row). Use `fields` to select which columns are returned; defaults to a lean subset.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args historyDealsArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		deals, err := c.GetHistoryDeals(ctx, accountID, args.TrdEnv, args.TrdMarket, args.BeginTime, args.EndTime, args.Codes)
		if err != nil {
			return toolError(fmt.Sprintf("get history deals: %v", err)), nil, nil
		}
		col, err := dealsToColumnar(deals, args.Fields)
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		return jsonResult(col), nil, nil
	})
}
