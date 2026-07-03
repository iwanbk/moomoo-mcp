package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

type securityGroupArgs struct {
	GroupType string `json:"group_type,omitempty"`
}

type userSecurityArgs struct {
	GroupName string `json:"group_name"`
}

// RegisterWatchlist registers the 2 watchlist read-only tools on the MCP server.
func RegisterWatchlist(s *mcp.Server, c moomoo.MoomooClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_user_security_group",
		Description: "List the user's watchlist groups. group_type: CUSTOM, SYSTEM, ALL (default ALL). Call this first to get the group_name needed by get_user_security.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args securityGroupArgs) (*mcp.CallToolResult, any, error) {
		groups, err := c.GetUserSecurityGroup(ctx, args.GroupType)
		if err != nil {
			return toolError(fmt.Sprintf("get user security group: %v", err)), nil, nil
		}
		return jsonResult(groups), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_user_security",
		Description: "Get the securities in one watchlist group. group_name comes from get_user_security_group.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args userSecurityArgs) (*mcp.CallToolResult, any, error) {
		securities, err := c.GetUserSecurity(ctx, args.GroupName)
		if err != nil {
			return toolError(fmt.Sprintf("get user security: %v", err)), nil, nil
		}
		return jsonResult(securities), nil, nil
	})
}
