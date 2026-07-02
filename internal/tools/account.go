package tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

// accountHeaderArgs identifies which trading account and market a request
// applies to. account_id is a decimal string (see moomoo.Account) so large
// IDs never lose precision when passed through JSON. trd_env defaults to
// SIMULATE when omitted.
type accountHeaderArgs struct {
	AccountID string `json:"account_id"`
	TrdEnv    string `json:"trd_env,omitempty"`
	TrdMarket string `json:"trd_market"`
}

func (a accountHeaderArgs) parseAccountID() (uint64, error) {
	id, err := strconv.ParseUint(a.AccountID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid account_id %q: %w", a.AccountID, err)
	}
	return id, nil
}

type getAccountsArgs struct{}

type assetsArgs struct {
	accountHeaderArgs
}

type positionsArgs struct {
	accountHeaderArgs
}

type accountSummaryArgs struct {
	accountHeaderArgs
}

type maxTradableArgs struct {
	accountHeaderArgs
	OrderType string  `json:"order_type"`
	Code      string  `json:"code"`
	Price     float64 `json:"price"`
}

type marginRatioArgs struct {
	accountHeaderArgs
	Codes []string `json:"codes"`
}

type cashFlowArgs struct {
	accountHeaderArgs
	ClearingDate string `json:"clearing_date"`
}

// RegisterAccount registers the 7 account read-only tools on the MCP server.
func RegisterAccount(s *mcp.Server, c moomoo.MoomooClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_accounts",
		Description: "List the trading accounts visible to this OpenD session, including account_id, trd_env (REAL/SIMULATE) and trd_markets. Call this first to get the account_id/trd_env/trd_market needed by the other account tools.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ getAccountsArgs) (*mcp.CallToolResult, any, error) {
		accs, err := c.GetAccounts(ctx)
		if err != nil {
			return toolError(fmt.Sprintf("get accounts: %v", err)), nil, nil
		}
		return jsonResult(accs), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_assets",
		Description: "Get funds/asset information (cash, buying power, market value) for one trading account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args assetsArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		assets, err := c.GetAssets(ctx, accountID, args.TrdEnv, args.TrdMarket)
		if err != nil {
			return toolError(fmt.Sprintf("get assets: %v", err)), nil, nil
		}
		return jsonResult(assets), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_positions",
		Description: "Get open positions for one trading account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args positionsArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		positions, err := c.GetPositions(ctx, accountID, args.TrdEnv, args.TrdMarket)
		if err != nil {
			return toolError(fmt.Sprintf("get positions: %v", err)), nil, nil
		}
		return jsonResult(positions), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_account_summary",
		Description: "Get a combined summary (assets + positions) for one trading account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args accountSummaryArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		assets, err := c.GetAssets(ctx, accountID, args.TrdEnv, args.TrdMarket)
		if err != nil {
			return toolError(fmt.Sprintf("get assets: %v", err)), nil, nil
		}
		positions, err := c.GetPositions(ctx, accountID, args.TrdEnv, args.TrdMarket)
		if err != nil {
			return toolError(fmt.Sprintf("get positions: %v", err)), nil, nil
		}
		return jsonResult(moomoo.AccountSummary{Assets: assets, Positions: positions}), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_max_tradable",
		Description: "Get the maximum tradable quantities (cash buy, margin buy, position sell, sell short) for a security under one account. order_type: NORMAL, MARKET, ABSOLUTE_LIMIT, AUCTION, AUCTION_LIMIT, SPECIAL_LIMIT, SPECIAL_LIMIT_ALL, STOP, STOP_LIMIT, MARKET_IF_TOUCHED, LIMIT_IF_TOUCHED, TRAILING_STOP, TRAILING_STOP_LIMIT.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args maxTradableArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		max, err := c.GetMaxTradable(ctx, accountID, args.TrdEnv, args.TrdMarket, args.OrderType, args.Code, args.Price)
		if err != nil {
			return toolError(fmt.Sprintf("get max tradable: %v", err)), nil, nil
		}
		return jsonResult(max), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_margin_ratio",
		Description: "Get margin ratio info (long/short permit, short pool remain, short fee rate) for one or more securities under one account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args marginRatioArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		ratios, err := c.GetMarginRatio(ctx, accountID, args.TrdEnv, args.TrdMarket, args.Codes)
		if err != nil {
			return toolError(fmt.Sprintf("get margin ratio: %v", err)), nil, nil
		}
		return jsonResult(ratios), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_cash_flow",
		Description: "Get the trading cash flow summary for one account on a given clearing date (format yyyyMMdd).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args cashFlowArgs) (*mcp.CallToolResult, any, error) {
		accountID, err := args.parseAccountID()
		if err != nil {
			return toolError(err.Error()), nil, nil
		}
		flow, err := c.GetCashFlow(ctx, accountID, args.TrdEnv, args.TrdMarket, args.ClearingDate)
		if err != nil {
			return toolError(fmt.Sprintf("get cash flow: %v", err)), nil, nil
		}
		return jsonResult(flow), nil, nil
	})
}
