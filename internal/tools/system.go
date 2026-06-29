package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iwanbk/moomoo-mcp/internal/moomoo"
)

type healthResult struct {
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

// RegisterSystem registers the check_health tool on the MCP server.
func RegisterSystem(s *mcp.Server, c moomoo.MoomooClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "check_health",
		Description: "Check connectivity to the OpenD gateway.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result := healthResult{Connected: true}
		if err := c.Health(ctx); err != nil {
			result.Connected = false
			result.Error = err.Error()
		}
		b, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
		}, nil, nil
	})
}
